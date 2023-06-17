package bot

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"main/src/commands"
	"os"
	"os/signal"
)

type DiscordBot struct {
	session  *discordgo.Session
	commands map[string]commands.BotCommand
}

func NewBot(apiToken string) *DiscordBot {
	token := "Bot" + apiToken
	dcClient, err := discordgo.New(token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	b := new(DiscordBot)
	b.session = dcClient

	return b
}

func (b *DiscordBot) RegisterCommand(command commands.BotCommand) {
	b.commands[command.Name()] = command
}

func (b *DiscordBot) Run() {
	err := b.session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	for name, cmd := range b.commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", cmd.Command())
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", name, err)
		}
	}

	defer func(client *discordgo.Session) {
		err := client.Close()
		if err != nil {
			log.Fatalf("error closing the discord session")
		}
	}(b.session)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}

func (b *DiscordBot) init() {
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if cmd, ok := b.commands[i.ApplicationCommandData().Name]; ok {
			cmd.Handle(s, i)
		}
	})
}
