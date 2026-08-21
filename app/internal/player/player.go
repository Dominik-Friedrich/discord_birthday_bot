// Package player implements the music player feature: queue YouTube tracks
// via /play, /togglepause, /stop, /skip. Audio streams straight from yt-dlp
// through ffmpeg into Discord, and each guild gets its own independent
// player.
package player

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
)

const featurePlayer = "featurePlayer"

const (
	embedColorQueued     = 0x1DB954
	embedColorNowPlaying = 0x5865F2

	// cleanupDelay is how long command confirmations (queued/paused/
	// stopped/skipped) stick around before being deleted, to keep the
	// channel tidy without yanking them away instantly. The "now playing"
	// announcement is exempt -- it's not tied to a command response, and is
	// meant to stay as a reference for what's currently playing.
	cleanupDelay = 5 * time.Minute
)

// scheduleCleanup deletes i's response after cleanupDelay.
func scheduleCleanup(s *discordgo.Session, i *discordgo.Interaction) {
	slog.Debug("scheduled message cleanup", "guild_id", i.GuildID, "delay", cleanupDelay)
	time.AfterFunc(cleanupDelay, func() {
		if err := s.InteractionResponseDelete(i); err != nil {
			slog.Warn("error cleaning up player message", "error", err)
			return
		}
		slog.Debug("cleaned up player message", "guild_id", i.GuildID)
	})
}

// TrackInfo is what a command needs to tell the user what got queued.
type TrackInfo struct {
	Title       string
	URL         string
	Thumbnail   string
	RequestedBy string // display name; empty to omit
}

// Embed builds a Discord embed announcing this track, its title linking to
// URL, with description as the caller-chosen status line (e.g. "Added to
// the queue", "Now playing").
func (t TrackInfo) Embed(description string, color int) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       t.Title,
		URL:         t.URL,
		Description: description,
		Color:       color,
	}
	if t.Thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: t.Thumbnail}
	}
	if t.RequestedBy != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: "Requested by " + t.RequestedBy}
	}
	return embed
}

// IPlayer is the playback control surface the slash commands drive, keeping
// the state machine decoupled from the Discord command layer.
type IPlayer interface {
	// Play returns whether the track started playing immediately (queue was
	// empty), so the caller can skip showing a redundant "queued" message --
	// the player announces "now playing" itself in that case.
	Play(ctx context.Context, i *discordgo.Interaction, query string) (track TrackInfo, startedImmediately bool, err error)
	Stop(i *discordgo.Interaction) error
	// TogglePause returns the state the player ended up in, so the command
	// can report whether it actually paused or resumed (or did nothing,
	// e.g. Stopped/Idle).
	TogglePause(i *discordgo.Interaction) (StateName, error)
	Forward(i *discordgo.Interaction, forwardCount uint) error
}

// Feature implements bot.Feature and owns one guildPlayer per Discord guild,
// created lazily on first use.
type Feature struct {
	session  *bot.Session
	resolver *resolver

	mu      sync.Mutex
	players map[string]*guildPlayer
}

func New(maxMediaDuration time.Duration) bot.Feature {
	return &Feature{
		resolver: newResolver(maxMediaDuration),
		players:  make(map[string]*guildPlayer),
	}
}

func (f *Feature) Init(session *bot.Session) error {
	f.session = session
	return nil
}

func (f *Feature) Name() string {
	return featurePlayer
}

func (f *Feature) Commands() []bot.Command {
	return []bot.Command{
		Play(f),
		Pause(f),
		Stop(f),
		Skip(f),
	}
}

func (f *Feature) Play(ctx context.Context, i *discordgo.Interaction, query string) (TrackInfo, bool, error) {
	p, err := f.playerFor(i.GuildID)
	if err != nil {
		return TrackInfo{}, false, err
	}
	return p.play(ctx, i, query)
}

func (f *Feature) Stop(i *discordgo.Interaction) error {
	p, err := f.playerFor(i.GuildID)
	if err != nil {
		return err
	}
	return p.stop()
}

func (f *Feature) TogglePause(i *discordgo.Interaction) (StateName, error) {
	p, err := f.playerFor(i.GuildID)
	if err != nil {
		return "", err
	}
	return p.togglePause()
}

func (f *Feature) Forward(i *discordgo.Interaction, forwardCount uint) error {
	p, err := f.playerFor(i.GuildID)
	if err != nil {
		return err
	}
	return p.forward(forwardCount)
}

// playerFor returns the guildPlayer for guildID, creating it on first use.
func (f *Feature) playerFor(guildID string) (*guildPlayer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if p, ok := f.players[guildID]; ok {
		return p, nil
	}

	slog.Debug("creating player for guild", "guild_id", guildID)
	p, err := newGuildPlayer(guildID, f.session, f.resolver)
	if err != nil {
		return nil, fmt.Errorf("initializing player: %w", err)
	}
	f.players[guildID] = p

	return p, nil
}
