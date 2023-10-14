package Player

import (
	"github.com/bwmarrin/dgvoice"
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
				Required:    true,
			},
		},
	}
}

func (p playCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	response := "success_playing"
	err := playAudio(s, i)
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

func playAudio(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	g, err := s.State.Guild(i.GuildID)

	// Look for the message sender in that guild's current voice states.
	var channelId string
	for _, vs := range g.VoiceStates {
		if vs.UserID == i.Member.User.ID {
			channelId = vs.ChannelID
			break
		}
	}

	if channelId == "" {
		// todo: error user not in channel
		return nil
	}

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(i.GuildID, channelId, false, true)
	if err != nil {
		// todo: error joining channel
		return nil
	}

	dgvoice.PlayAudioFile(vc, "./resources/EIyixC9NsLI.opus", make(chan bool))

	// Disconnect from the provided voice channel.
	err = vc.Disconnect()
	if err != nil {
		// todo: error disconnecting channel
		return nil
	}

	vc.Close()
	return err
}

func (p playCommand) Name() string {
	return play
}
