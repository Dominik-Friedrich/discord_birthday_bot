package features

import (
	"github.com/bwmarrin/discordgo"
	"main/src/commands"
)

type BotFeature interface {
	Init(session *discordgo.Session) error
	Name() string
	Commands() []commands.BotCommand
}
