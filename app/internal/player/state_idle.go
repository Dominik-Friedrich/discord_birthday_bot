package player

import "github.com/bwmarrin/discordgo"

// stateIdle means "connected, queue is empty" -- advanceQueue only enters
// here when there's nothing to skip, so Forward is a no-op.
type stateIdle struct {
	player *guildPlayer
}

func (s stateIdle) State() StateName {
	return Idle
}

func (s stateIdle) OnEntry(_ State) { return }
func (s stateIdle) OnExit()         { return }

func (s stateIdle) Play(i *discordgo.Interaction, item queueItem) error {
	s.player.enqueueBack(item)
	s.player.advanceQueue()

	return nil
}

func (s stateIdle) Stop() error {
	s.player.states.setState(Stopped)
	return nil
}

func (s stateIdle) TogglePause() error {
	return nil
}

func (s stateIdle) Forward(uint) error {
	return nil
}
