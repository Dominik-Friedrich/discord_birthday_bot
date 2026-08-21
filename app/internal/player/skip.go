package player

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
)

const (
	skipCommandName = "skip"
	paramSkipAmount = "skip_amount"
)

type skipCommand struct {
	player IPlayer
}

func Skip(player IPlayer) bot.Command {
	return &skipCommand{player: player}
}

func (p *skipCommand) Name() string {
	return skipCommandName
}

func (p *skipCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(
		discordgo.PermissionViewChannel |
			discordgo.PermissionVoiceConnect |
			discordgo.PermissionVoiceSpeak |
			discordgo.PermissionSendMessages,
	)

	return &discordgo.ApplicationCommand{
		Name:                     skipCommandName,
		Description:              "Skip the currently playing song",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        paramSkipAmount,
				Description: "The number of tracks to skip",
				MinValue:    new(1.0),
			},
		},
	}
}

func (p *skipCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	skipAmount := skipAmountOption(i)

	response := skipMessage(skipAmount)
	if err := p.player.Forward(i.Interaction, skipAmount); err != nil {
		slog.Warn("error skipping audio", "error", err)
		response = err.Error()
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	}); err != nil {
		slog.Warn("error responding to command prompt", "error", err)
	}
}

func skipAmountOption(i *discordgo.InteractionCreate) uint {
	optionMap := bot.OptionMap(i.ApplicationCommandData().Options)

	skipAmount := uint(1)
	if opt, ok := optionMap[paramSkipAmount]; ok {
		skipAmount = uint(opt.UintValue())
	}
	return skipAmount
}

func skipMessage(skipAmount uint) string {
	if skipAmount == 1 {
		return "Skipped to the next track."
	}
	return fmt.Sprintf("Skipped %d tracks.", skipAmount)
}
