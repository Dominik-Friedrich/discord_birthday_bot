package commands

import "github.com/bwmarrin/discordgo"

type BotCommand interface {
	Command() *discordgo.ApplicationCommand
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
	Name() string
}
