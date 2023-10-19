package Player

import "github.com/bwmarrin/discordgo"

type IPlayer interface {
	Play(interaction *discordgo.Interaction, mediaName string) error
	Stop() error
	TogglePause() error
	Forward(forwardCount uint) error
	Backward() error
	Playing() bool
}
