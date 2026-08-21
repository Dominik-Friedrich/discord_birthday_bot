package bot

import (
	"context"
	"fmt"
	"log/slog"

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

func New(apiToken, applicationId string) (*DiscordBot, error) {
	token := "Bot " + apiToken
	dcClient, err := discordgo.New(token)
	if err != nil {
		return nil, fmt.Errorf("creating discord client: %w", err)
	}

	b := new(DiscordBot)
	b.session = &Session{
		ApplicationId: applicationId,
		Session:       dcClient,
	}
	b.commands = make(map[string]Command)
	b.features = make(map[string]Feature)
	b.init()

	return b, nil
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
// blocks until ctx is cancelled.
func (b *DiscordBot) Run(ctx context.Context) error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("opening discord session: %w", err)
	}
	defer func(session *Session) {
		if err := session.Close(); err != nil {
			slog.Warn("error closing the discord session", "error", err)
		}
	}(b.session)

	for _, feature := range b.features {
		if err := feature.Init(b.session); err != nil {
			return fmt.Errorf("initializing feature %q: %w", feature.Name(), err)
		}
	}

	slog.Info("adding commands")
	commands := make([]*discordgo.ApplicationCommand, 0, len(b.commands))
	for name, cmd := range b.commands {
		commands = append(commands, cmd.Command())
		slog.Debug("added command", "name", name)
	}
	if _, err := b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, "", commands); err != nil {
		return fmt.Errorf("registering commands: %w", err)
	}

	slog.Info("bot is running")
	<-ctx.Done()

	slog.Info("gracefully shutting down")
	return nil
}

func (b *DiscordBot) init() {
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if cmd, ok := b.commands[i.ApplicationCommandData().Name]; ok {
			cmd.Handle(s, i)
		}
	})
}
