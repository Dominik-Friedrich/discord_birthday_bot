package Player

import (
	"errors"
	"github.com/bwmarrin/discordgo"
)

type stateIdle struct {
	player *player
}

func (s stateIdle) State() StateName {
	return Idle
}

func (s stateIdle) OnEntry(oldState State) {
	s.player.playNextMedia()
}

func (s stateIdle) OnExit() {
}

func (s stateIdle) Play(i *discordgo.Interaction, mediaName string) error {
	err := s.player.AddQueueBack(mediaName)
	if err != nil {
		return err
	}

	s.player.playNextMedia()
	return nil
}

func (s stateIdle) Stop() error {
	s.player.states.setState(Stopped)

	return nil
}

func (s stateIdle) TogglePause() error {
	return nil
}

func (s stateIdle) Forward() error {
	//TODO implement me
	return errors.New("unimplemented feature")
}

func (s stateIdle) Backward() error {
	//TODO implement me
	return errors.New("unimplemented feature")
}

func (s stateIdle) Playing() bool {
	return false
}
