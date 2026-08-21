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
	state, err := p.player.TogglePause(i.Interaction)

	response := togglePauseMessage(state)
	if err != nil {
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
		return
	}
	scheduleCleanup(s, i.Interaction)
}

func togglePauseMessage(state StateName) string {
	switch state {
	case Paused:
		return "Paused."
	case Playing:
		return "Resumed."
	default:
		return "Nothing is playing right now."
	}
}
