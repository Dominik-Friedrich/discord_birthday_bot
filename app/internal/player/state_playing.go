package player

import "github.com/bwmarrin/discordgo"

type statePlaying struct {
	player *guildPlayer
}

func (s statePlaying) State() StateName {
	return Playing
}

// OnEntry only re-establishes the speaking indicator -- it does not start
// playback itself. Starting a track is advanceQueue's job, done explicitly,
// because setState no-ops when the state *name* doesn't change (e.g. one
// track finishing and the next starting are both just "Playing"), so an
// OnEntry-driven start would never fire for the second track onward.
func (s statePlaying) OnEntry(_ State) {
	s.player.speaking(true)
}

func (s statePlaying) OnExit() {
	s.player.speaking(false)
}

func (s statePlaying) Play(i *discordgo.Interaction, item queueItem) error {
	if err := s.player.initVc(i); err != nil {
		return err
	}

	s.player.enqueueBack(item)
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

// Forward drops the skipped tracks and jumps to whatever's next (or Idle)
// without waiting for the current track's stop to be confirmed.
func (s statePlaying) Forward(forwardCount uint) error {
	if forwardCount == 0 {
		return nil
	}

	const currentlyPlayingOffset = 1
	s.player.removeQueueFront(forwardCount - currentlyPlayingOffset)
	s.player.dcPlayer.Stop()
	s.player.advanceQueue()

	return nil
}
