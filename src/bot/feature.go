package bot

type Feature interface {
	Init(session *Session) error
	Name() string
	Commands() []Command
}
