package Player

import (
	"errors"
	"github.com/bwmarrin/discordgo"
)

type statePaused struct {
	player *player
}

func (s statePaused) State() StateName {
	return Paused
}

func (s statePaused) OnEntry(oldState State) {
	s.player.dcPlayer.Pause()
}

func (s statePaused) OnExit() {
	s.player.dcPlayer.Unpause()
}

func (s statePaused) Play(i *discordgo.Interaction, mediaName string) error {
	return s.player.AddQueueBack(mediaName)
}

func (s statePaused) Stop() error {
	s.player.states.setState(Stopped)
	return nil
}

func (s statePaused) TogglePause() error {
	s.player.states.setState(Playing)
	return nil
}

func (s statePaused) Forward(forwardCount uint) error {
	const currentlyPlayingOffset = 1
	s.player.RemoveQueueFront(forwardCount - currentlyPlayingOffset)
	s.player.dcPlayer.Stop()
	s.player.states.setState(Idle)

	return nil
}

func (s statePaused) Backward() error {
	//TODO implement me
	return errors.New("unimplemented feature")
}

func (s statePaused) Playing() bool {
	return false
}
