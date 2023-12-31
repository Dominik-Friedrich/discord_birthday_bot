package commands

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
	"main/src/repository/birthday"
)

const (
	removeBirthday = "remove-birthday"
)

type removeBirthdayCommand struct {
	birthdays birthday.Repository
}

func RemoveBirthday(repo birthday.Repository) bot.Command {
	cmd := new(removeBirthdayCommand)
	cmd.birthdays = repo
	return cmd
}

func (a *removeBirthdayCommand) Name() string {
	return removeBirthday
}

func (a *removeBirthdayCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionManageRoles)

	return &discordgo.ApplicationCommand{
		Name:                     removeBirthday,
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
		log.Println(err.Error())
		response = err.Error()
	} else {
		err := a.birthdays.RemoveBirthday(birthdayUser)
		if err != nil {
			log.Println(log.WARN, err.Error())
			response = "something went horribly wrong D:"
		}
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

func (a *removeBirthdayCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (birthday.User, error) {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var errs error
	var birthdayUser birthday.User

	if option, ok := optionMap[paramUser]; ok {
		usr := option.UserValue(s)
		birthdayUser.UserId = usr.ID
		birthdayUser.UserName = usr.Username
	} else {
		errs = errors.Join(errors.New("you need to specify the birthday user"))
	}

	return birthdayUser, errs
}
