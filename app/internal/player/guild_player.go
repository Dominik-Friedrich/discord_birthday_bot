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
	query       string
	result      goutubedl.Result
	requestedBy string
}

// trackInfo builds the TrackInfo shown in both the queued and now-playing
// messages.
func (item queueItem) trackInfo() TrackInfo {
	return TrackInfo{
		Title:       item.result.Info.Title,
		URL:         item.result.Info.WebpageURL,
		Thumbnail:   item.result.Info.Thumbnail,
		RequestedBy: item.requestedBy,
	}
}

// playResult is what controlLoop reports back once it applies a play
// request: whether it started immediately, and any error from doing so.
type playResult struct {
	startedImmediately bool
	err                error
}

type playRequest struct {
	interaction *discordgo.Interaction
	item        queueItem
	// resp, if set, receives the outcome once controlLoop applies it.
	resp chan playResult
}

// togglePauseResult is what controlLoop reports back for a toggle-pause
// request: the state it landed in, and any error applying it.
type togglePauseResult struct {
	state StateName
	err   error
}

type forwardRequest struct {
	count uint
	resp  chan error
}

// guildPlayer drives playback for a single Discord guild: one voice
// connection, one queue, one state machine, one control-loop goroutine.
type guildPlayer struct {
	guildID  string
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

	// announceChannelID is the text channel "now playing" messages get
	// posted to -- the channel /play was last used from. Only ever touched
	// from controlLoop, so it needs no lock of its own.
	announceChannelID string

	// pendingAdvance marks a Stop() sent for a skip: once the audio player
	// confirms the old track stopped, controlLoop starts the next one
	// itself, rather than advanceQueue running synchronously right after
	// Stop() returns. Only ever touched from controlLoop.
	pendingAdvance bool

	playCh        chan playRequest
	stopCh        chan chan error
	togglePauseCh chan chan togglePauseResult
	forwardCh     chan forwardRequest
}

func newGuildPlayer(guildID string, session *bot.Session, resolver *resolver) (*guildPlayer, error) {
	p := &guildPlayer{
		guildID:  guildID,
		session:  session,
		resolver: resolver,

		playCh:        make(chan playRequest),
		stopCh:        make(chan chan error),
		togglePauseCh: make(chan chan togglePauseResult),
		forwardCh:     make(chan forwardRequest),

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

func (p *guildPlayer) play(ctx context.Context, i *discordgo.Interaction, query string) (TrackInfo, bool, error) {
	result, err := p.resolver.resolve(ctx, query)
	if err != nil {
		return TrackInfo{}, false, err
	}

	item := queueItem{query: query, result: result, requestedBy: requesterName(i)}

	resp := make(chan playResult, 1)
	p.playCh <- playRequest{interaction: i, item: item, resp: resp}
	res := <-resp
	if res.err != nil {
		return TrackInfo{}, false, res.err
	}

	return item.trackInfo(), res.startedImmediately, nil
}

func (p *guildPlayer) stop() error {
	resp := make(chan error, 1)
	p.stopCh <- resp
	return <-resp
}

// togglePause waits for controlLoop to actually apply the toggle and reports
// the state it landed in, so the caller can tell a pause from a resume (or a
// no-op, e.g. nothing was playing).
func (p *guildPlayer) togglePause() (StateName, error) {
	resp := make(chan togglePauseResult, 1)
	p.togglePauseCh <- resp
	res := <-resp
	return res.state, res.err
}

func (p *guildPlayer) forward(forwardCount uint) error {
	resp := make(chan error, 1)
	p.forwardCh <- forwardRequest{count: forwardCount, resp: resp}
	return <-resp
}

// controlLoop is the only goroutine allowed to touch the state machine, so
// transitions always happen one at a time.
func (p *guildPlayer) controlLoop() {
	for {
		var err error
		select {
		case req := <-p.playCh:
			startedImmediately := p.states.getState().State() == Stopped || p.states.getState().State() == Idle
			slog.Debug("play request received", "guild_id", p.guildID, "query", req.item.query, "requested_by", req.item.requestedBy, "state", p.states.getState().State())
			p.announceChannelID = req.interaction.ChannelID
			err = p.states.getState().Play(req.interaction, req.item)
			if req.resp != nil {
				req.resp <- playResult{startedImmediately: startedImmediately && err == nil, err: err}
			}
		case resp := <-p.togglePauseCh:
			slog.Debug("toggle-pause request received", "guild_id", p.guildID, "state", p.states.getState().State())
			err = p.states.getState().TogglePause()
			resp <- togglePauseResult{state: p.states.getState().State(), err: err}
		case resp := <-p.stopCh:
			slog.Debug("stop request received", "guild_id", p.guildID, "state", p.states.getState().State())
			err = p.states.getState().Stop()
			resp <- err
		case req := <-p.forwardCh:
			slog.Debug("forward request received", "guild_id", p.guildID, "count", req.count, "state", p.states.getState().State())
			err = p.states.getState().Forward(req.count)
			req.resp <- err
		case ctx := <-p.dcPlayDoneChan:
			slog.Debug("audio player reported done", "guild_id", p.guildID, "media", p.currentItem.query, "exit_reason", ctx.ExitReason, "pending_advance", p.pendingAdvance)
			switch ctx.ExitReason {
			case discord.Finished:
				p.advanceQueue()
			case discord.Error:
				slog.Warn("playback error, disconnecting", "guild_id", p.guildID, "media", p.currentItem.query)
				p.pendingAdvance = false
				p.states.setState(Stopped)
			case discord.Stopped:
				// Set only by a skip; a plain /stop leaves this false.
				if p.pendingAdvance {
					p.pendingAdvance = false
					p.advanceQueue()
				}
			}
		}

		if err != nil {
			slog.Warn("error in player control loop", "guild_id", p.guildID, "error", err)
		}
	}
}

// advanceQueue pops the next queued item and starts it, or moves to Idle if
// the queue is empty. startPlayback/announceNowPlaying are called explicitly
// here rather than from Playing's OnEntry, since setState no-ops when the
// state name doesn't change (one track ending into the next is still just
// "Playing").
func (p *guildPlayer) advanceQueue() {
	p.queueMutex.Lock()
	if p.queue.Len() == 0 {
		p.queueMutex.Unlock()
		slog.Debug("queue empty, going idle", "guild_id", p.guildID)
		p.states.setState(Idle)
		return
	}
	item := p.queue.PopFront()
	remaining := p.queue.Len()
	p.queueMutex.Unlock()

	slog.Debug("advancing to next track", "guild_id", p.guildID, "query", item.query, "remaining_in_queue", remaining)
	p.currentItem = item
	p.states.setState(Playing)
	p.startPlayback()
	p.announceNowPlaying()
}

// startPlayback hands the current item to the audio player; it returns as
// soon as the audio player accepts it, not when playback finishes.
func (p *guildPlayer) startPlayback() {
	p.vcMutex.Lock()
	vc := p.currentVc
	p.vcMutex.Unlock()

	slog.Debug("handing track to audio player", "guild_id", p.guildID, "title", p.currentItem.result.Info.Title)
	p.dcPlayer.Play(discord.PlayerContext{
		Vc:     vc,
		Result: p.currentItem.result,
	})
}

// announceNowPlaying posts a "now playing" embed for the current item to
// the last channel /play was used from. The post itself runs off controlLoop
// so a slow/rate-limited Discord API call can't stall command processing for
// the whole guild.
func (p *guildPlayer) announceNowPlaying() {
	if p.announceChannelID == "" {
		slog.Debug("no announce channel set, skipping now-playing message", "guild_id", p.guildID)
		return
	}

	track := p.currentItem.trackInfo()
	channelID := p.announceChannelID
	guildID := p.guildID
	go func() {
		if _, err := p.session.ChannelMessageSendEmbed(channelID, track.Embed("Now playing", embedColorNowPlaying)); err != nil {
			slog.Warn("error announcing now playing", "guild_id", guildID, "error", err)
			return
		}
		slog.Debug("posted now-playing message", "guild_id", guildID, "channel_id", channelID, "title", track.Title)
	}()
}

// requesterName is the display name to credit in "Requested by" -- the
// member's nickname if they have one, else their username.
func requesterName(i *discordgo.Interaction) string {
	if i.Member == nil || i.Member.User == nil {
		return ""
	}
	if i.Member.Nick != "" {
		return i.Member.Nick
	}
	return i.Member.User.Username
}

func (p *guildPlayer) enqueueBack(item queueItem) {
	p.queueMutex.Lock()
	p.queue.PushBack(item)
	size := p.queue.Len()
	p.queueMutex.Unlock()

	slog.Debug("enqueued track at back", "guild_id", p.guildID, "query", item.query, "queue_size", size)
}

func (p *guildPlayer) enqueueFront(item queueItem) {
	p.queueMutex.Lock()
	p.queue.PushFront(item)
	size := p.queue.Len()
	p.queueMutex.Unlock()

	slog.Debug("enqueued track at front", "guild_id", p.guildID, "query", item.query, "queue_size", size)
}

func (p *guildPlayer) removeQueueFront(count uint) {
	p.queueMutex.Lock()
	defer p.queueMutex.Unlock()

	removed := uint(0)
	for range count {
		if p.queue.Len() == 0 {
			break
		}
		p.queue.PopFront()
		removed++
	}
	slog.Debug("removed tracks from queue front", "guild_id", p.guildID, "requested", count, "removed", removed, "queue_size", p.queue.Len())
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

	slog.Debug("joining voice channel", "guild_id", p.guildID, "channel_id", channelID)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vc, err := p.session.ChannelVoiceJoin(ctx, i.GuildID, channelID, false, true)
	if err != nil {
		return fmt.Errorf("joining voice channel: %w", err)
	}

	// Discord requires DAVE (E2EE) for voice; frames sent before the DAVE
	// handshake completes go out unencrypted and get dropped, so wait for it
	// here rather than losing the first moment of audio on every track.
	slog.Debug("waiting for DAVE encryption to be ready", "guild_id", p.guildID)
	if err := vc.WaitForDAVEReady(ctx); err != nil {
		return fmt.Errorf("waiting for voice encryption: %w", err)
	}

	slog.Debug("voice channel ready", "guild_id", p.guildID, "channel_id", channelID)
	p.currentVc = vc
	return nil
}

func (p *guildPlayer) speaking(b bool) {
	p.vcMutex.Lock()
	defer p.vcMutex.Unlock()

	if p.currentVc == nil {
		return
	}
	slog.Debug("setting speaking status", "guild_id", p.guildID, "speaking", b)
	if err := p.currentVc.Speaking(b); err != nil {
		slog.Warn("error setting speaking status", "guild_id", p.guildID, "error", err)
	}
}
