package discord

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"io"
	"layeh.com/gopus"
	"os/exec"
	"strconv"
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

type PlayerContext struct {
	Vc         *discordgo.VoiceConnection
	MediaName  string
	ExitReason ExitReason
}

type AudioPlayer struct {
	playDone chan<- PlayerContext
	play     chan PlayerContext
	stop     chan struct{}
	pause    chan struct{}
	unpause  chan struct{}

	opusEncoder *gopus.Encoder
}

func NewPlayer(playDone chan<- PlayerContext) (*AudioPlayer, error) {
	p := new(AudioPlayer)
	p.playDone = playDone
	p.play = make(chan PlayerContext)
	p.stop = make(chan struct{})
	p.pause = make(chan struct{})
	p.unpause = make(chan struct{})

	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		return nil, fmt.Errorf("error initializing player: %s", err)
	}
	p.opusEncoder = opusEncoder

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

func (p *AudioPlayer) asyncPlayRoutine() {
	for {
		select {
		case ctx := <-p.play:
			p.playAudioFile(ctx.Vc, ctx.MediaName)
			p.playDone <- ctx
		}
	}
}

// PlayAudioFile will play the given filename to the already connected
// Discord voice server/channel.  voice websocket and udp socket
// must already be setup before this will work.
//
// copied and pasted from dgvoice (adjusted)
func (p *AudioPlayer) playAudioFile(v *discordgo.VoiceConnection, filename string) ExitReason {
	// Create a shell command "object" to run.
	run := exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	ffmpegout, err := run.StdoutPipe()
	if err != nil {
		dgvoice.OnError("StdoutPipe Error", err)
		return Error
	}

	ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384)

	// Starts the ffmpeg command
	err = run.Start()
	if err != nil {
		dgvoice.OnError("RunStart Error", err)
		return Error
	}

	// prevent memory leak from residual ffmpeg streams
	defer run.Process.Kill()

	send := make(chan []int16, 2)

	doneChan := make(chan bool)
	go func() {
		defer close(doneChan)
		p.SendPCM(v, send)
	}()

	// Producer Goroutine

	defer close(send)
	for {
		select {
		case <-doneChan:
			return Finished
		case <-p.stop:
			log.Println(log.INFO, "player stopped")
			return Stopped
		case <-p.pause:
			<-p.unpause

		default:
			// read data from ffmpeg stdout
			audiobuf := make([]int16, frameSize*channels)
			err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return Finished
			}
			if err != nil {
				dgvoice.OnError("error reading from ffmpeg stdout", err)
				return Error
			}
			// Send received PCM to the sendPCM channel
			send <- audiobuf
		}
	}
}

// SendPCM will receive on the provied channel encode
// received PCM data into Opus then send that to Discordgo
//
// copied and pasted from dgvoice
func (p *AudioPlayer) SendPCM(v *discordgo.VoiceConnection, pcm <-chan []int16) {
	if pcm == nil {
		return
	}

	for {
		// read pcm from chan, exit if channel is closed.
		recv, ok := <-pcm
		if !ok {
			dgvoice.OnError("PCM Channel closed", nil)
			return
		}

		// try encoding pcm frame with Opus
		opus, err := p.opusEncoder.Encode(recv, frameSize, maxBytes)
		if err != nil {
			dgvoice.OnError("Encoding Error", err)
			return
		}

		if v.Ready == false || v.OpusSend == nil {
			// OnError(fmt.Sprintf("Discordgo not ready for opus packets. %+v : %+v", v.Ready, v.OpusSend), nil)
			// Sending errors here might not be suited
			return
		}
		// send encoded opus data to the sendOpus channel
		v.OpusSend <- opus
	}
}
