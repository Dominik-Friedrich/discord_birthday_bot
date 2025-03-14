package commands

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
	"main/src/repository"
)

const (
	complainReply = "add-complaint-reply"
	paramReply    = "reply"
)

type complainReplyCommand struct {
	repo    repository.Repository
	replies *Cache
}

func AddComplainReply(repo repository.Repository, replies *Cache) bot.Command {
	cmd := new(complainReplyCommand)
	cmd.repo = repo
	cmd.replies = replies

	return cmd
}

func (a *complainReplyCommand) Name() string {
	return complainReply
}

func (a *complainReplyCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionSendMessages)

	return &discordgo.ApplicationCommand{
		Name:                     a.Name(),
		Description:              "Adds a complaint reply",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        paramReply,
				Description: "The reply the bot can use",
				Required:    true,
			},
		},
	}
}

func (a *complainReplyCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	newComplaint, err := a.validateUserInput(s, i)

	var response string
	if err != nil {
		log.Println(err.Error())
		response = err.Error()
	} else {
		err := a.repo.AddComplaintReply(&newComplaint)
		if err != nil {
			log.Println(log.WARN, err.Error())
		}
		response = "Added complaint reply"
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		log.Println("error responding to command prompt", err.Error())
	}
}

func (a *complainReplyCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (repository.Reply, error) {
	options := i.ApplicationCommandData().Options
	if i.Member == nil {
		return repository.Reply{}, errors.New("unable to determine complainant")
	}

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var err error
	newComplaintReply := repository.Reply{
		GuildId: i.GuildID,
		User: &repository.User{
			GuildId:  i.GuildID,
			UserId:   i.Member.User.ID,
			Username: i.Member.User.Username,
		},
		Text: "",
	}
	if i.Member.Nick != "" {
		newComplaintReply.User.Nickname = &i.Member.Nick
	}

	if opt, ok := optionMap[paramReply]; ok {
		complaintText := opt.StringValue()
		newComplaintReply.Text = complaintText
	} else {
		err = errors.New("reply can't be empty")
	}

	return newComplaintReply, err
}

func (a *complainReplyCommand) renewCache() error {
	replies, err := a.repo.GetComplaintReplies()

	if err != nil {
		a.replies.Refresh(replies)
	}

	return err
}
