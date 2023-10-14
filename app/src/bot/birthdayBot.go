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

	log.Println("Adding commands...")
	for name, cmd := range b.commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", cmd.Command())
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", name, err)
		}
		log.Println("Added command: ", cmd.Name())
	}

	defer func(client *Session) {
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
