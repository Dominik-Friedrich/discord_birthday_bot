package Player

type PlayerState int

const (
	Stopped PlayerState = 0
	Paused  PlayerState = 1
	Playing PlayerState = 2
)
