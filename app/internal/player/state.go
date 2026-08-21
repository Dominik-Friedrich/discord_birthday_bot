package player

import "github.com/bwmarrin/discordgo"

const (
	Stopped StateName = "stopped"
	Idle    StateName = "idle"
	Playing StateName = "playing"
	Paused  StateName = "paused"
)

type StateName string

// State is one state in the player's state machine; each state decides what
// a control action means (e.g. Play joins+starts in stateStopped, just
// queues in statePlaying).
type State interface {
	OnEntry(oldState State)
	OnExit()
	State() StateName

	Play(i *discordgo.Interaction, item queueItem) error
	Stop() error
	TogglePause() error
	Forward(forwardCount uint) error
}
