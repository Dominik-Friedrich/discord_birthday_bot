package commands

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
	"main/src/repository/birthday"
	"main/src/repository/complaint"
	"math/rand"
	"time"
)

const (
	complain       = "complain"
	paramUser      = "user"
	paramComplaint = "complaint"

	screamsInPain = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
)

type complainCommand struct {
	repo    complaint.Repository
	replies *Cache
}

func Complain(repo complaint.Repository, replies *Cache) bot.Command {
	cmd := new(complainCommand)
	cmd.repo = repo
	cmd.replies = replies

	return cmd
}

func (a *complainCommand) Name() string {
	return complain
}

func (a *complainCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionManageRoles)

	return &discordgo.ApplicationCommand{
		Name:                     complain,
		Description:              "Adds the birthday of a user",
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

	response := screamsInPain
	if err != nil {
		log.Println(err.Error())
		response = err.Error()
	} else {
		err := a.repo.AddComplaint(newComplaint)
		if err != nil {
			log.Println(log.WARN, err.Error())
			response = "you can't even write a complaint?"
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

func (a *complainCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (complaint.Complaint, error) {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Or convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var err error
	var newComplaint complaint.Complaint

	if option, ok := optionMap[paramUser]; ok {
		usr := option.UserValue(s)
		newComplaint.User = &birthday.User{
			GuildId:  i.GuildID,
			UserId:   usr.ID,
			UserName: usr.Username,
			Birthday: time.Time{},
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
	a.replies.Lock()
	defer a.replies.Unlock()

	if !a.replies.Valid() {
		err := a.renewCache()
		if err != nil {
			log.Println(log.WARN, "could not refresh reply cache: ", err.Error())
		}
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
