package commands

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
	"main/src/repository"
	"math/rand"
	"time"
)

const (
	complain       = "complain"
	paramUser      = "user"
	paramComplaint = "complaint"
)

type complainCommand struct {
	repo    repository.Repository
	replies *Cache
}

func Complain(repo repository.Repository, replies *Cache) bot.Command {
	cmd := new(complainCommand)
	cmd.repo = repo
	cmd.replies = replies

	return cmd
}

func (a *complainCommand) Name() string {
	return complain
}

func (a *complainCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionSendMessages)

	return &discordgo.ApplicationCommand{
		Name:                     complain,
		Description:              "Complain about something",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        paramComplaint,
				Description: "What are you complaining about?",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        paramUser,
				Description: "User to complain about",
			},
		},
	}
}

func (a *complainCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	newComplaint, err := a.validateUserInput(s, i)

	var response string
	if err != nil {
		log.Println(err.Error())
		response = err.Error()
	} else {
		err := a.repo.AddComplaint(newComplaint)
		if err != nil {
			log.Println(log.WARN, err.Error())
		}
	}

	response = a.randomReply()

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// Ignore type for now, they will be discussed in "responses"
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		log.Println("error responding to command prompt", err.Error())
	}
}

func (a *complainCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (repository.Complaint, error) {
	options := i.ApplicationCommandData().Options
	if i.Member == nil {
		return repository.Complaint{}, errors.New("unable to determine complainant")
	}

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var err error
	newComplaint := repository.Complaint{
		Complainant: &repository.User{
			GuildId:  i.GuildID,
			UserId:   i.Member.User.ID,
			Username: i.Member.User.Username,
		},
	}
	if i.Member.Nick != "" {
		newComplaint.Complainant.Nickname = &i.Member.Nick
	}

	if option, ok := optionMap[paramUser]; ok {
		usr := option.UserValue(s)
		newComplaint.AgainstUser = &repository.User{
			GuildId:  i.GuildID,
			UserId:   usr.ID,
			Username: usr.Username,
			Birthday: time.Time{},
		}
		if i.Member.Nick != "" {
			newComplaint.AgainstUser.Nickname = &i.Member.Nick
		}
	}

	if opt, ok := optionMap[paramComplaint]; ok {
		complaintText := opt.StringValue()
		newComplaint.Text = complaintText
	} else {
		err = errors.New("stop complaining about nothing")
	}

	return newComplaint, err
}

func (a *complainCommand) randomReply() string {
	const screamsInPain = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	a.replies.Lock()
	defer a.replies.Unlock()

	if !a.replies.Valid() {
		err := a.renewCache()
		if err != nil {
			log.Println(log.WARN, "could not refresh reply cache: ", err.Error())
		}
	}

	if a.replies.Len() == 0 {
		return screamsInPain
	}

	index := rand.Intn(a.replies.Len())

	complaintReply, _ := a.replies.Get(index)

	return complaintReply.Text
}

func (a *complainCommand) renewCache() error {
	replies, err := a.repo.GetComplaintReplies()

	if err != nil {
		a.replies.Refresh(replies)
	}

	return err
}
