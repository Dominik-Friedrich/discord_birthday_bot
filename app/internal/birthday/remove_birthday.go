package birthday

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/user"

	"github.com/bwmarrin/discordgo"
)

const removeBirthdayCommandName = "remove-birthday"

type removeBirthdayCommand struct {
	users UserRepository
}

func RemoveBirthday(users UserRepository) bot.Command {
	return &removeBirthdayCommand{users: users}
}

func (a *removeBirthdayCommand) Name() string {
	return removeBirthdayCommandName
}

func (a *removeBirthdayCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionManageRoles)

	return &discordgo.ApplicationCommand{
		Name:                     removeBirthdayCommandName,
		Description:              "Removes the birthday of a user",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        paramUser,
				Description: "User of which the birthday is to be removed",
				Required:    true,
			},
		},
	}
}

func (a *removeBirthdayCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	birthdayUser, err := a.validateUserInput(s, i)

	response := "successfully removed the birthday!"
	if err != nil {
		slog.Warn("invalid remove-birthday input", "error", err)
		response = err.Error()
	} else if err := a.users.RemoveBirthday(context.Background(), birthdayUser); err != nil {
		slog.Warn("error removing birthday", "error", err)
		response = "something went horribly wrong D:"
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

func (a *removeBirthdayCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (user.User, error) {
	optionMap := bot.OptionMap(i.ApplicationCommandData().Options)

	var birthdayUser user.User
	option, ok := optionMap[paramUser]
	if !ok {
		return birthdayUser, errors.New("you need to specify the birthday user")
	}

	usr := option.UserValue(s)
	birthdayUser.UserId = usr.ID
	birthdayUser.Username = usr.Username

	return birthdayUser, nil
}
