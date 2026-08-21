package player

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
)

type stateStopped struct {
	player *guildPlayer
}

func (s stateStopped) State() StateName {
	return Stopped
}

// OnEntry disconnects and closes the voice channel if one is open, and stops
// any currently playing media.
func (s stateStopped) OnEntry(_ State) {
	// A skip that was in flight when /stop landed shouldn't cause a track
	// to start once this Stop() is confirmed.
	s.player.pendingAdvance = false
	s.player.dcPlayer.Stop()

	s.player.vcMutex.Lock()
	vc := s.player.currentVc
	if vc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := vc.Disconnect(ctx); err != nil {
			slog.Warn("error disconnecting from voice channel", "error", err)
		}
		cancel()
		s.player.currentVc = nil
	}
	s.player.vcMutex.Unlock()
}

func (s stateStopped) OnExit() { return }

func (s stateStopped) Play(i *discordgo.Interaction, item queueItem) error {
	if err := s.player.initVc(i); err != nil {
		return err
	}

	s.player.enqueueFront(item)
	s.player.advanceQueue()

	return nil
}

func (s stateStopped) Stop() error {
	return nil
}

func (s stateStopped) TogglePause() error {
	return nil
}

func (s stateStopped) Forward(uint) error {
	return nil
}
