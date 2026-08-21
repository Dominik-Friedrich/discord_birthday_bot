package bot

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
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

func New(apiToken, applicationId string) *DiscordBot {
	token := "Bot " + apiToken
	dcClient, err := discordgo.New(token)
	if err != nil {
		slog.Error("invalid bot parameters", "error", err)
		os.Exit(1)
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

// Run opens the Discord session, initializes all registered features, and
// blocks until an interrupt signal is received.
func (b *DiscordBot) Run() {
	err := b.session.Open()
	if err != nil {
		slog.Error("cannot open the session", "error", err)
		os.Exit(1)
	}
	defer func(client *Session) {
		if err := client.Close(); err != nil {
			slog.Error("error closing the discord session", "error", err)
		}
	}(b.session)

	for _, feature := range b.features {
		if err := feature.Init(b.session); err != nil {
			slog.Error("error registering feature", "feature", feature.Name(), "error", err)
			os.Exit(1)
		}
	}

	slog.Info("adding commands")
	commands := make([]*discordgo.ApplicationCommand, 0, len(b.commands))
	for name, cmd := range b.commands {
		commands = append(commands, cmd.Command())
		slog.Debug("added command", "name", name)
	}
	if _, err := b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, "", commands); err != nil {
		slog.Error("error creating commands", "error", err)
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	slog.Info("bot is running, press Ctrl+C to exit")
	<-ctx.Done()

	slog.Info("gracefully shutting down")
}

func (b *DiscordBot) init() {
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if cmd, ok := b.commands[i.ApplicationCommandData().Name]; ok {
			cmd.Handle(s, i)
		}
	})
}
