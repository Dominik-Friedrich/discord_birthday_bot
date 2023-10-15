package Player

type State int

const (
	Stopped State = iota
	Paused
	Playing
	Idle
)

func (s State) String() string {
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
