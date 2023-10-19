package Player

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
)

const (
	skipAmountParam = "skip_amount"
)

type skipCommand struct {
	name   string
	player IPlayer
}

func Skip(player IPlayer) bot.Command {
	cmd := new(skipCommand)
	cmd.player = player
	cmd.name = "skip"

	return cmd
}

func (p skipCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(
		discordgo.PermissionViewChannel |
			discordgo.PermissionVoiceConnect |
			discordgo.PermissionVoiceSpeak |
			discordgo.PermissionSendMessages,
	)

	return &discordgo.ApplicationCommand{
		Name:                     p.name,
		Description:              "Skip the currently playing song",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        skipAmountParam,
				Description: "The number of tracks to skip",
			},
		},
	}
}

func (p skipCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	response := "success_skipping"
	err := p.skipAudio(s, i)
	if err != nil {
		log.Println(log.WARN, "error skipping audio: ", err.Error())
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

func (p skipCommand) skipAudio(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Or convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	skipAmount := uint(1)
	if opt, ok := optionMap[skipAmountParam]; ok {
		skipAmount = uint(opt.UintValue())
	}

	err := p.player.Forward(skipAmount)
	return err
}

func (p skipCommand) Name() string {
	return p.name
}
