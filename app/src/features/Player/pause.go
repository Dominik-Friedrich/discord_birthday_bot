package Player

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
)

const (
	pause = "pause"
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
		Name:                     pause,
		Description:              "Pauses the player",
		DefaultMemberPermissions: &neededPermissions,
	}
}

func (p pauseCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	response := "success_paused"
	p.player.Pause()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
	return pause
}
