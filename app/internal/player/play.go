package player

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
)

const (
	playCommandName = "play"
	paramQuery      = "query"
)

type playCommand struct {
	player IPlayer
}

func Play(player IPlayer) bot.Command {
	return &playCommand{player: player}
}

func (p *playCommand) Name() string {
	return playCommandName
}

func (p *playCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(
		discordgo.PermissionViewChannel |
			discordgo.PermissionVoiceConnect |
			discordgo.PermissionVoiceSpeak |
			discordgo.PermissionSendMessages,
	)

	return &discordgo.ApplicationCommand{
		Name:                     playCommandName,
		Description:              "Play a song either from a URL or search.",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        paramQuery,
				Description: "The query to search for.",
				Required:    true,
			},
		},
	}
}

func (p *playCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Resolving can take a few seconds -- defer to avoid missing Discord's
	// 3-second response window.
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		slog.Warn("error deferring interaction response", "error", err)
		return
	}

	response, err := p.playAudio(s, i)
	if err != nil {
		slog.Warn("error playing audio", "error", err)
		response = err.Error()
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	}); err != nil {
		slog.Warn("error editing interaction response", "error", err)
	}
}

func (p *playCommand) playAudio(_ *discordgo.Session, i *discordgo.InteractionCreate) (string, error) {
	if i.Member == nil {
		return "", errors.New("unable to determine who to play for")
	}

	optionMap := bot.OptionMap(i.ApplicationCommandData().Options)
	opt, ok := optionMap[paramQuery]
	if !ok {
		return "", errors.New("query is a required field")
	}

	title, err := p.player.Play(context.Background(), i.Interaction, opt.StringValue())
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("queued: %s", title), nil
}
