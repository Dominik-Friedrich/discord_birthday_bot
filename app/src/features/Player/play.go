package Player

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
)

const (
	param = "query"
)

type playCommand struct {
	name   string
	player IPlayer
}

func Play(player IPlayer) bot.Command {
	cmd := new(playCommand)
	cmd.player = player
	cmd.name = "play"

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
		Name:                     p.name,
		Description:              "Play a song either from a URL or search.",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        param,
				Description: "The query to search for.",
				Required:    true,
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
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Or convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var query string
	if opt, ok := optionMap[param]; ok {
		query = opt.StringValue()
	} else {
		return errors.New("query is a required field")
	}

	err := p.player.Play(i.Interaction, query)
	return err
}

func (p playCommand) Name() string {
	return p.name
}
