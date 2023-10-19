package Player

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
)

type statePlaying struct {
	player *player
}

func (s statePlaying) State() StateName {
	return Playing
}

func (s statePlaying) OnEntry(oldState State) {
	s.player.speaking(true)
}

func (s statePlaying) OnExit() {
	err := s.player.AddHistoryFront(s.player.currentMedia)
	if err != nil {
		log.Println(log.WARN, "error adding media to history: ", err)
	}

	s.player.speaking(false)
}

func (s statePlaying) Play(i *discordgo.Interaction, mediaName string) error {
	err := s.player.initVc(i)
	if err != nil {
		return err
	}

	err = s.player.AddQueueBack(mediaName)
	if err != nil {
		return err
	}

	return nil
}

func (s statePlaying) Stop() error {
	s.player.states.setState(Stopped)

	return nil
}

func (s statePlaying) TogglePause() error {
	s.player.states.setState(Paused)
	return nil
}

func (s statePlaying) Forward() error {
	//TODO implement me
	return errors.New("unimplemented feature")
}

func (s statePlaying) Backward() error {
	//TODO implement me
	return errors.New("unimplemented feature")
}

func (s statePlaying) Playing() bool {
	return false
}
