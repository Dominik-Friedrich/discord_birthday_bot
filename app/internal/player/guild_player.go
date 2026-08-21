package player

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gammazero/deque"
	"github.com/wader/goutubedl"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/player/discord"
)

// queueItem is a track already resolved and duration-validated by resolver.
type queueItem struct {
	query  string
	result goutubedl.Result
}

type playRequest struct {
	interaction *discordgo.Interaction
	item        queueItem
}

// guildPlayer drives playback for a single Discord guild: one voice
// connection, one queue, one state machine, one control-loop goroutine.
type guildPlayer struct {
	session  *bot.Session
	resolver *resolver

	currentVc *discordgo.VoiceConnection
	vcMutex   sync.Mutex

	dcPlayer       *discord.AudioPlayer
	dcPlayDoneChan chan discord.PlayerContext

	states *StateMachine

	currentItem queueItem
	queue       deque.Deque[queueItem]
	queueMutex  sync.Mutex

	playCh        chan playRequest
	stopCh        chan struct{}
	togglePauseCh chan struct{}
	forwardCh     chan uint
}

func newGuildPlayer(session *bot.Session, resolver *resolver) (*guildPlayer, error) {
	p := &guildPlayer{
		session:  session,
		resolver: resolver,

		playCh:        make(chan playRequest),
		stopCh:        make(chan struct{}),
		togglePauseCh: make(chan struct{}),
		forwardCh:     make(chan uint),

		dcPlayDoneChan: make(chan discord.PlayerContext),
	}

	dcPlayer, err := discord.NewPlayer(p.dcPlayDoneChan)
	if err != nil {
		return nil, fmt.Errorf("initializing audio player: %w", err)
	}
	p.dcPlayer = dcPlayer

	p.states = NewStateMachine(p)

	go p.controlLoop()

	return p, nil
}

func (p *guildPlayer) play(ctx context.Context, i *discordgo.Interaction, query string) (string, error) {
	result, err := p.resolver.resolve(ctx, query)
	if err != nil {
		return "", err
	}

	p.playCh <- playRequest{interaction: i, item: queueItem{query: query, result: result}}

	return result.Info.Title, nil
}

func (p *guildPlayer) stop() error {
	p.stopCh <- struct{}{}
	return nil
}

func (p *guildPlayer) togglePause() error {
	p.togglePauseCh <- struct{}{}
	return nil
}

func (p *guildPlayer) forward(forwardCount uint) error {
	p.forwardCh <- forwardCount
	return nil
}

// controlLoop is the only goroutine allowed to touch the state machine, so
// transitions always happen one at a time.
func (p *guildPlayer) controlLoop() {
	for {
		var err error
		select {
		case req := <-p.playCh:
			err = p.states.getState().Play(req.interaction, req.item)
		case <-p.togglePauseCh:
			err = p.states.getState().TogglePause()
		case <-p.stopCh:
			err = p.states.getState().Stop()
		case forwardCount := <-p.forwardCh:
			err = p.states.getState().Forward(forwardCount)
		case ctx := <-p.dcPlayDoneChan:
			switch ctx.ExitReason {
			case discord.Finished:
				p.advanceQueue()
			case discord.Error:
				slog.Warn("playback error, disconnecting", "media", p.currentItem.query)
				p.states.setState(Stopped)
			case discord.Stopped:
				// caller that issued the stop drives the next transition.
			}
		}

		if err != nil {
			slog.Warn("error in player control loop", "error", err)
		}
	}
}

// advanceQueue pops the next queued item and moves to Playing, or to Idle
// if the queue is empty.
func (p *guildPlayer) advanceQueue() {
	p.queueMutex.Lock()
	if p.queue.Len() == 0 {
		p.queueMutex.Unlock()
		p.states.setState(Idle)
		return
	}
	item := p.queue.PopFront()
	p.queueMutex.Unlock()

	p.currentItem = item
	p.states.setState(Playing)
}

// startPlayback hands the current item to the audio player; it returns as
// soon as the audio player accepts it, not when playback finishes.
func (p *guildPlayer) startPlayback() {
	p.vcMutex.Lock()
	vc := p.currentVc
	p.vcMutex.Unlock()

	p.dcPlayer.Play(discord.PlayerContext{
		Vc:     vc,
		Result: p.currentItem.result,
	})
}

func (p *guildPlayer) enqueueBack(item queueItem) {
	p.queueMutex.Lock()
	p.queue.PushBack(item)
	p.queueMutex.Unlock()
}

func (p *guildPlayer) enqueueFront(item queueItem) {
	p.queueMutex.Lock()
	p.queue.PushFront(item)
	p.queueMutex.Unlock()
}

func (p *guildPlayer) removeQueueFront(count uint) {
	p.queueMutex.Lock()
	defer p.queueMutex.Unlock()

	for range count {
		if p.queue.Len() == 0 {
			return
		}
		p.queue.PopFront()
	}
}

// initVc joins the interacting member's voice channel if not already
// connected. It never moves channels once connected.
func (p *guildPlayer) initVc(i *discordgo.Interaction) error {
	p.vcMutex.Lock()
	defer p.vcMutex.Unlock()

	if p.currentVc != nil {
		return nil
	}

	guild, err := p.session.State.Guild(i.GuildID)
	if err != nil {
		return fmt.Errorf("looking up guild: %w", err)
	}

	var channelID string
	for _, vs := range guild.VoiceStates {
		if vs.UserID == i.Member.User.ID {
			channelID = vs.ChannelID
			break
		}
	}
	if channelID == "" {
		return errors.New("you need to be in a voice channel")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vc, err := p.session.ChannelVoiceJoin(ctx, i.GuildID, channelID, false, true)
	if err != nil {
		return fmt.Errorf("joining voice channel: %w", err)
	}

	// Discord requires DAVE (E2EE) for voice; frames sent before the DAVE
	// handshake completes go out unencrypted and get dropped, so wait for it
	// here rather than losing the first moment of audio on every track.
	if err := vc.WaitForDAVEReady(ctx); err != nil {
		return fmt.Errorf("waiting for voice encryption: %w", err)
	}

	p.currentVc = vc
	return nil
}

func (p *guildPlayer) speaking(b bool) {
	p.vcMutex.Lock()
	defer p.vcMutex.Unlock()

	if p.currentVc == nil {
		return
	}
	if err := p.currentVc.Speaking(b); err != nil {
		slog.Warn("error setting speaking status", "error", err)
	}
}
