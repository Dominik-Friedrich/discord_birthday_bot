package player

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
)

const togglePauseCommandName = "togglepause"

type pauseCommand struct {
	player IPlayer
}

func Pause(player IPlayer) bot.Command {
	return &pauseCommand{player: player}
}

func (p *pauseCommand) Name() string {
	return togglePauseCommandName
}

func (p *pauseCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(
		discordgo.PermissionViewChannel |
			discordgo.PermissionVoiceConnect |
			discordgo.PermissionVoiceSpeak |
			discordgo.PermissionSendMessages,
	)

	return &discordgo.ApplicationCommand{
		Name:                     togglePauseCommandName,
		Description:              "Pause/unpauses the player",
		DefaultMemberPermissions: &neededPermissions,
	}
}

func (p *pauseCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	response := "success_paused"
	if err := p.player.TogglePause(i.Interaction); err != nil {
		slog.Warn("error toggling pause", "error", err)
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
