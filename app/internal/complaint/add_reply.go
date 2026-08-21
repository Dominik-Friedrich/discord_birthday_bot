package complaint

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/user"

	"github.com/bwmarrin/discordgo"
)

const (
	addComplainReplyCommandName = "add-complaint-reply"
	paramReply                  = "reply"
)

type complainReplyCommand struct {
	repo    *Repository
	replies *cache
}

func AddComplainReply(repo *Repository, replies *cache) bot.Command {
	return &complainReplyCommand{repo: repo, replies: replies}
}

func (a *complainReplyCommand) Name() string {
	return addComplainReplyCommandName
}

func (a *complainReplyCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionManageMessages)

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
	ctx := context.Background()
	newReply, err := a.validateUserInput(s, i)

	var response string
	if err != nil {
		slog.Warn("invalid add-complaint-reply input", "error", err)
		response = err.Error()
	} else if err := a.repo.AddComplaintReply(ctx, &newReply); err != nil {
		slog.Warn("error adding complaint reply", "error", err)
		response = "something went horribly wrong D:"
	} else {
		a.replies.Lock()
		if err := refreshCache(ctx, a.repo, a.replies); err != nil {
			slog.Warn("could not refresh reply cache", "error", err)
		}
		a.replies.Unlock()
		response = "Added complaint reply"
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

func (a *complainReplyCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (Reply, error) {
	if i.Member == nil {
		return Reply{}, errors.New("unable to determine complainant")
	}
	optionMap := bot.OptionMap(i.ApplicationCommandData().Options)

	newReply := Reply{
		GuildId: i.GuildID,
		User: &user.User{
			GuildId:  i.GuildID,
			UserId:   i.Member.User.ID,
			Username: i.Member.User.Username,
		},
	}
	if i.Member.Nick != "" {
		newReply.User.Nickname = &i.Member.Nick
	}

	opt, ok := optionMap[paramReply]
	if !ok {
		return newReply, errors.New("reply can't be empty")
	}
	newReply.Text = opt.StringValue()

	return newReply, nil
}
