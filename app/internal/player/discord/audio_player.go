// Package discord streams a resolved track through ffmpeg and Opus-encodes
// it into a Discord voice connection. It knows nothing about queues, guilds,
// or the player's state machine.
package discord

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/wader/goutubedl"
	"layeh.com/gopus"
)

const (
	channels  int = 2     // 1 for mono, 2 for stereo
	frameRate int = 48000 // audio sampling rate
	frameSize int = 960   // uint16 size of each audio frame
	maxBytes      = frameSize * 2 * 2
)

type ExitReason int

const (
	Finished ExitReason = iota
	Stopped
	Error
)

func (e ExitReason) String() string {
	switch e {
	case Finished:
		return "finished"
	case Stopped:
		return "stopped"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

type PlayerContext struct {
	Vc         *discordgo.VoiceConnection
	Result     goutubedl.Result
	ExitReason ExitReason
}

// AudioPlayer streams one track at a time: yt-dlp into ffmpeg, ffmpeg's PCM
// output Opus-encoded and sent to Discord.
type AudioPlayer struct {
	playDone chan<- PlayerContext
	play     chan PlayerContext
	stop     chan struct{}
	pause    chan struct{}
	unpause  chan struct{}

	opusEncoder *gopus.Encoder
}

func NewPlayer(playDone chan<- PlayerContext) (*AudioPlayer, error) {
	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		return nil, fmt.Errorf("initializing opus encoder: %w", err)
	}

	p := &AudioPlayer{
		playDone: playDone,
		play:     make(chan PlayerContext),
		// Buffered so Stop/Pause/Unpause never block on send: asyncPlayRoutine's
		// playDone send (below) isn't select-guarded, so a caller racing that
		// window would otherwise deadlock waiting for a receiver that's itself
		// waiting on the caller.
		stop:        make(chan struct{}, 1),
		pause:       make(chan struct{}, 1),
		unpause:     make(chan struct{}, 1),
		opusEncoder: opusEncoder,
	}

	go p.asyncPlayRoutine()

	return p, nil
}

func (p *AudioPlayer) Play(ctx PlayerContext) {
	p.play <- ctx
}

func (p *AudioPlayer) Stop() {
	p.stop <- struct{}{}
}

func (p *AudioPlayer) Pause() {
	p.pause <- struct{}{}
}

func (p *AudioPlayer) Unpause() {
	p.unpause <- struct{}{}
}

// asyncPlayRoutine also drains stop/pause/unpause while idle, so those calls
// are always safe even when nothing is playing.
func (p *AudioPlayer) asyncPlayRoutine() {
	for {
		select {
		case ctx := <-p.play:
			ctx.ExitReason = p.playAudioStream(ctx.Vc, ctx.Result)
			p.playDone <- ctx
		case <-p.stop:
		case <-p.pause:
		case <-p.unpause:
		}
	}
}

// playAudioStream streams the track live through ffmpeg for PCM transcoding
// and sends it to v. Cancelling ctx kills both the yt-dlp and ffmpeg
// subprocesses, since both are tied to it.
func (p *AudioPlayer) playAudioStream(v *discordgo.VoiceConnection, result goutubedl.Result) (reason ExitReason) {
	title := result.Info.Title
	slog.Debug("starting audio stream", "title", title)
	defer func() { slog.Debug("audio stream ended", "title", title, "exit_reason", reason) }()

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := result.Download(ctx, "bestaudio")
	if err != nil {
		slog.Warn("error starting audio stream", "title", title, "error", err)
		cancel()
		return Error
	}
	defer stream.Close()
	defer cancel()

	run := exec.CommandContext(ctx, "ffmpeg", "-i", "pipe:0", "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	run.Stdin = stream

	var stderr bytes.Buffer
	run.Stderr = &stderr

	ffmpegOut, err := run.StdoutPipe()
	if err != nil {
		slog.Warn("error creating ffmpeg stdout pipe", "title", title, "error", err)
		return Error
	}
	ffmpegBuf := bufio.NewReaderSize(ffmpegOut, 16384)

	if err := run.Start(); err != nil {
		slog.Warn("error starting ffmpeg", "title", title, "error", err)
		return Error
	}
	slog.Debug("ffmpeg started", "title", title)

	send := make(chan []int16, 2)
	sendResult := make(chan ExitReason, 1)
	go func() {
		sendResult <- p.sendPCM(v, send)
	}()
	defer close(send)

	for {
		select {
		case reason := <-sendResult:
			return reason
		case <-p.stop:
			return Stopped
		case <-p.pause:
			select {
			case <-p.stop:
				return Stopped
			case <-p.unpause:
			}
		default:
			audiobuf := make([]int16, frameSize*channels)
			readErr := binary.Read(ffmpegBuf, binary.LittleEndian, &audiobuf)
			if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
				return Finished
			}
			if readErr != nil {
				slog.Warn("error reading from ffmpeg stdout", "error", readErr, "ffmpeg_stderr", stderr.String())
				return Error
			}
			// select-guarded so a dead sendPCM (exited on error) can't block
			// this forever.
			select {
			case send <- audiobuf:
			case reason := <-sendResult:
				return reason
			case <-p.stop:
				return Stopped
			}
		}
	}
}

// sendPCM Opus-encodes and sends PCM frames until pcm closes (Finished) or
// encoding/sending fails (Error) -- a not-ready connection used to fail
// silently here instead.
func (p *AudioPlayer) sendPCM(v *discordgo.VoiceConnection, pcm <-chan []int16) (reason ExitReason) {
	// v.Status/v.OpusSend are read without the connection's own lock, so a
	// concurrent Kill() closing OpusSend can make the send below panic --
	// recovered here so that fails this one track instead of the process.
	defer func() {
		if r := recover(); r != nil {
			slog.Warn("recovered panic sending opus frame", "panic", r)
			reason = Error
		}
	}()

	for recv := range pcm {
		opus, err := p.opusEncoder.Encode(recv, frameSize, maxBytes)
		if err != nil {
			slog.Warn("error encoding opus frame", "error", err)
			return Error
		}

		if v.Status != discordgo.VoiceConnectionStatusReady || v.OpusSend == nil {
			slog.Warn("voice connection not ready, aborting playback")
			return Error
		}
		v.OpusSend <- opus
	}
	return Finished
}
