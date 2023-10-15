package Player

const (
	Stopped StateName = "stopped"
	Paused  StateName = "paused"
	Playing StateName = "playing"
	Idle    StateName = "idle"
)

type State interface {
	OnEntry(oldState State)
	OnExit()
	State() StateName

	IPlayer
}

type StateName string
