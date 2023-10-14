package Player

type IPlayer interface {
	Play(mediaName string) error
	Stop() error
	Pause() error
	Forward() error
	Backward() error
	Playing() bool
}

type WebPlayer interface {
	IPlayer
	SupportedSites() []string
}
