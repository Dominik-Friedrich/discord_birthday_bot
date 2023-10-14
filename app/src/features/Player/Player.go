package Player

import "github.com/bwmarrin/discordgo"

type IPlayer interface {
	Play(interaction *discordgo.Interaction, mediaName string) error
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
