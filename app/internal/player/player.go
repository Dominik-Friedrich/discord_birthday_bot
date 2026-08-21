// Package player implements the music player feature: queue YouTube tracks
// via /play, /togglepause, /stop, /skip. Audio streams straight from yt-dlp
// through ffmpeg into Discord, and each guild gets its own independent
// player.
package player

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
)

const featurePlayer = "featurePlayer"

// IPlayer is the playback control surface the slash commands drive, keeping
// the state machine decoupled from the Discord command layer.
type IPlayer interface {
	Play(ctx context.Context, i *discordgo.Interaction, query string) (title string, err error)
	Stop(i *discordgo.Interaction) error
	TogglePause(i *discordgo.Interaction) error
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

func (f *Feature) Play(ctx context.Context, i *discordgo.Interaction, query string) (string, error) {
	p, err := f.playerFor(i.GuildID)
	if err != nil {
		return "", err
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

func (f *Feature) TogglePause(i *discordgo.Interaction) error {
	p, err := f.playerFor(i.GuildID)
	if err != nil {
		return err
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

	p, err := newGuildPlayer(f.session, f.resolver)
	if err != nil {
		return nil, fmt.Errorf("initializing player: %w", err)
	}
	f.players[guildID] = p

	return p, nil
}
