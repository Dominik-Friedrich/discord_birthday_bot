package birthday

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/user"

	"github.com/bwmarrin/discordgo"
)

const (
	addBirthdayCommandName = "add-birthday"
	paramUser              = "user"
	paramBirthday          = "birthday"
	birthdayFormat         = "02/01"
	birthdayFormatReadable = "DD/MM"
)

type addBirthdayCommand struct {
	users          UserRepository
	userAddedEvent chan user.User
}

func AddBirthday(users UserRepository, userAddedEvent chan user.User) bot.Command {
	return &addBirthdayCommand{
		users:          users,
		userAddedEvent: userAddedEvent,
	}
}

func (a *addBirthdayCommand) Name() string {
	return addBirthdayCommandName
}

func (a *addBirthdayCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionManageRoles)

	return &discordgo.ApplicationCommand{
		Name:                     addBirthdayCommandName,
		Description:              "Adds the birthday of a user",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        paramUser,
				Description: "User of which the birthday is to be added",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        paramBirthday,
				Description: fmt.Sprintf("Birthday date of the user in format '%s'", birthdayFormatReadable),
				Required:    true,
			},
		},
	}
}

func (a *addBirthdayCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	birthdayUser, err := a.validateUserInput(s, i)

	response := "successfully added the birthday!"
	if err != nil {
		slog.Warn("invalid add-birthday input", "error", err)
		response = err.Error()
	} else {
		if err := a.users.UpsertUser(context.Background(), &birthdayUser); err != nil {
			slog.Warn("error upserting user", "error", err)
			response = "something went horribly wrong D:"
		} else if a.userAddedEvent != nil {
			a.userAddedEvent <- birthdayUser
		}
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

func (a *addBirthdayCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (user.User, error) {
	optionMap := bot.OptionMap(i.ApplicationCommandData().Options)

	var errs error
	var birthdayUser user.User

	if option, ok := optionMap[paramUser]; ok {
		usr := option.UserValue(s)

		member, err := s.GuildMember(i.GuildID, usr.ID)
		if err != nil {
			return user.User{}, err
		}
		birthdayUser.UserId = member.User.ID
		birthdayUser.GuildId = i.GuildID
		birthdayUser.Username = member.User.Username
		if member.Nick != "" {
			birthdayUser.Nickname = &member.Nick
		}
	} else {
		errs = errors.Join(errs, errors.New("you need to specify the birthday user"))
	}

	if opt, ok := optionMap[paramBirthday]; ok {
		birthdayString := opt.StringValue()
		birthdayDate, err := time.Parse(birthdayFormat, birthdayString)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("the birthday has to be in the format '%s'", birthdayFormatReadable))
		}
		birthdayUser.Birthday = birthdayDate
	} else {
		errs = errors.Join(errs, fmt.Errorf("you need to specify the birthday date. Format '%s'", birthdayFormatReadable))
	}

	return birthdayUser, errs
}
