package Player

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
)

const (
	stop = "stop"
)

type stopCommand struct {
	player IPlayer
}

func Stop(player IPlayer) bot.Command {
	cmd := new(stopCommand)
	cmd.player = player

	return cmd
}

func (p stopCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(
		discordgo.PermissionViewChannel |
			discordgo.PermissionVoiceConnect |
			discordgo.PermissionVoiceSpeak |
			discordgo.PermissionSendMessages,
	)

	return &discordgo.ApplicationCommand{
		Name:                     stop,
		Description:              "Stop the player",
		DefaultMemberPermissions: &neededPermissions,
	}
}

func (p stopCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	response := "success_stopped"
	err := p.player.Stop()
	if err != nil {
		response = "error_stopped"
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

func (p stopCommand) Name() string {
	return stop
}
