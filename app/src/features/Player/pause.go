package Player

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
)

const (
	togglepause = "togglepause"
)

type pauseCommand struct {
	player IPlayer
}

func Pause(player IPlayer) bot.Command {
	cmd := new(pauseCommand)
	cmd.player = player

	return cmd
}

func (p pauseCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(
		discordgo.PermissionViewChannel |
			discordgo.PermissionVoiceConnect |
			discordgo.PermissionVoiceSpeak |
			discordgo.PermissionSendMessages,
	)

	return &discordgo.ApplicationCommand{
		Name:                     togglepause,
		Description:              "Pause/Unpauses the player",
		DefaultMemberPermissions: &neededPermissions,
	}
}

func (p pauseCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	response := "success_paused"
	err := p.player.TogglePause()
	if err != nil {
		response = "error_pausing"
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		log.Println("error responding to command prompt: ", err.Error())
	}
}

func (p pauseCommand) Name() string {
	return togglepause
}
