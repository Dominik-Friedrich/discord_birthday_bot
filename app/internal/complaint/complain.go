package complaint

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/user"

	"github.com/bwmarrin/discordgo"
)

const (
	complainCommandName = "complain"
	paramUser           = "user"
	paramComplaint      = "complaint"
)

type complainCommand struct {
	repo    *Repository
	replies *cache
}

func Complain(repo *Repository, replies *cache) bot.Command {
	return &complainCommand{repo: repo, replies: replies}
}

func (a *complainCommand) Name() string {
	return complainCommandName
}

func (a *complainCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionSendMessages)

	return &discordgo.ApplicationCommand{
		Name:                     complainCommandName,
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
	ctx := context.Background()
	newComplaint, err := a.validateUserInput(s, i)

	var response string
	if err != nil {
		slog.Warn("invalid complain input", "error", err)
		response = err.Error()
	} else {
		if err := a.repo.AddComplaint(ctx, newComplaint); err != nil {
			slog.Warn("error adding complaint", "error", err)
		}
		response = a.randomReply(ctx)
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	}); err != nil {
		slog.Warn("error responding to command prompt", "error", err)
	}
}

func (a *complainCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (Complaint, error) {
	if i.Member == nil {
		return Complaint{}, errors.New("unable to determine complainant")
	}
	optionMap := bot.OptionMap(i.ApplicationCommandData().Options)

	var err error
	newComplaint := Complaint{
		Complainant: &user.User{
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
		newComplaint.AgainstUser = &user.User{
			GuildId:  i.GuildID,
			UserId:   usr.ID,
			Username: usr.Username,
		}
		if i.Member.Nick != "" {
			newComplaint.AgainstUser.Nickname = &i.Member.Nick
		}
	}

	if opt, ok := optionMap[paramComplaint]; ok {
		newComplaint.Text = opt.StringValue()
	} else {
		err = errors.New("stop complaining about nothing")
	}

	return newComplaint, err
}

func (a *complainCommand) randomReply(ctx context.Context) string {
	const screamsInPain = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	a.replies.Lock()
	defer a.replies.Unlock()

	if !a.replies.Valid() {
		if err := refreshCache(ctx, a.repo, a.replies); err != nil {
			slog.Warn("could not refresh reply cache", "error", err)
		}
	}

	if a.replies.Len() == 0 {
		return screamsInPain
	}

	index := rand.IntN(a.replies.Len())

	complaintReply, _ := a.replies.Get(index)

	return complaintReply.Text
}
