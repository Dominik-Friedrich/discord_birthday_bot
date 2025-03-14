package bot

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"os"
	"os/signal"
)

type DiscordBot struct {
	session  *Session
	commands map[string]Command
	features map[string]Feature
}

type Session struct {
	ApplicationId string
	*discordgo.Session
}

func NewBot(apiToken, applicationId string) *DiscordBot {
	token := "Bot " + apiToken
	dcClient, err := discordgo.New(token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	b := new(DiscordBot)
	b.session = &Session{
		ApplicationId: applicationId,
		Session:       dcClient,
	}
	b.commands = make(map[string]Command)
	b.features = make(map[string]Feature)
	b.init()

	return b
}

func (b *DiscordBot) RegisterCommand(command Command) {
	b.commands[command.Name()] = command
}

func (b *DiscordBot) RegisterFeature(feature Feature) {
	b.features[feature.Name()] = feature

	for _, command := range feature.Commands() {
		b.commands[command.Name()] = command
	}
}

func (b *DiscordBot) Session() *Session {
	return b.session
}

func (b *DiscordBot) Run() {
	err := b.session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	for _, feature := range b.features {
		err := feature.Init(b.session)
		if err != nil {
			log.Fatalf("Error registerung feature '%s': %v", feature.Name(), err)
		}
	}

	log.Println(log.INFO, "Adding commands...")
	commands := make([]*discordgo.ApplicationCommand, 0, len(b.commands))
	for name, cmd := range b.commands {
		commands = append(commands, cmd.Command())
		log.Println(log.INFO, "Added commands: ", name)
	}
	_, err = b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, "", commands)
	if err != nil {
		log.Panicf("Error creating commands: %v", err)
	}

	defer func(client *Session) {
		err := client.Close()
		if err != nil {
			log.Fatalf("error closing the discord session")
		}
	}(b.session)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println(log.INFO, "Press Ctrl+C to exit")
	<-stop

	log.Println(log.INFO, "Gracefully shutting down.")
}

func (b *DiscordBot) init() {
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if cmd, ok := b.commands[i.ApplicationCommandData().Name]; ok {
			cmd.Handle(s, i)
		}
	})
}
