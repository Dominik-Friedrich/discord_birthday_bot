package player

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
)

const stopCommandName = "stop"

type stopCommand struct {
	player IPlayer
}

func Stop(player IPlayer) bot.Command {
	return &stopCommand{player: player}
}

func (p *stopCommand) Name() string {
	return stopCommandName
}

func (p *stopCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(
		discordgo.PermissionViewChannel |
			discordgo.PermissionVoiceConnect |
			discordgo.PermissionVoiceSpeak |
			discordgo.PermissionSendMessages,
	)

	return &discordgo.ApplicationCommand{
		Name:                     stopCommandName,
		Description:              "Stop the player",
		DefaultMemberPermissions: &neededPermissions,
	}
}

func (p *stopCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	response := "success_stopped"
	if err := p.player.Stop(i.Interaction); err != nil {
		slog.Warn("error stopping player", "error", err)
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
