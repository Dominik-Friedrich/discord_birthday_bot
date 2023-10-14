package Player

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
)

const (
	play  = "play"
	param = "url"
)

type playCommand struct {
	player IPlayer
}

func Play(player IPlayer) bot.Command {
	cmd := new(playCommand)
	cmd.player = player

	return cmd
}

func (p playCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(
		discordgo.PermissionViewChannel |
			discordgo.PermissionVoiceConnect |
			discordgo.PermissionVoiceSpeak |
			discordgo.PermissionSendMessages,
	)

	return &discordgo.ApplicationCommand{
		Name:                     play,
		Description:              "Play something",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        param,
				Description: "URL to play audio from. Also supports search queries",
			},
		},
	}
}

func (p playCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	response := "success_playing"
	err := p.playAudio(s, i)
	if err != nil {
		log.Println(log.WARN, "error playing audio: ", err.Error())
		response = err.Error()
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

func (p playCommand) playAudio(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := p.player.Play(i.Interaction, "./resources/EIyixC9NsLI.opus")
	return err
}

func (p playCommand) Name() string {
	return play
}
