package Player

type PlayerState int

const (
	Stopped PlayerState = 0
	Paused  PlayerState = 1
	Playing PlayerState = 2
)

func (s PlayerState) String() string {
	switch s {
	case Stopped:
		return "stopped"
	case Paused:
		return "paused"
	case Playing:
		return "playing"
	}

	return "unkown status"
}
