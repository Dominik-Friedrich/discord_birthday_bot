package player

import (
	"log/slog"

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
	s.player.dcPlayer.Stop()

	s.player.vcMutex.Lock()
	vc := s.player.currentVc
	if vc != nil {
		if err := vc.Disconnect(); err != nil {
			slog.Warn("error disconnecting from voice channel", "error", err)
		}
		vc.Close()
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
