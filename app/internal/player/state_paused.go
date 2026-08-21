package player

import "github.com/bwmarrin/discordgo"

type statePaused struct {
	player *guildPlayer
}

func (s statePaused) State() StateName {
	return Paused
}

func (s statePaused) OnEntry(_ State) {
	s.player.dcPlayer.Pause()
}

func (s statePaused) OnExit() {
	s.player.dcPlayer.Unpause()
}

func (s statePaused) Play(_ *discordgo.Interaction, item queueItem) error {
	s.player.enqueueBack(item)
	return nil
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
	if forwardCount == 0 {
		return nil
	}

	const currentlyPlayingOffset = 1
	s.player.removeQueueFront(forwardCount - currentlyPlayingOffset)
	s.player.dcPlayer.Stop()
	s.player.advanceQueue()

	return nil
}
