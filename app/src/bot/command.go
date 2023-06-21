package bot

import "github.com/bwmarrin/discordgo"

type Command interface {
	Command() *discordgo.ApplicationCommand
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
	Name() string
}
