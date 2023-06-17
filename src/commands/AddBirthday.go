package commands

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"main/src/repository"
	"time"
)

const (
	addBirthday            = "add-birthday"
	paramUser              = "user"
	paramBirthday          = "birthday"
	birthdayFormat         = "01/02"
	birthdayFormatReadable = "DD/MM"
)

type addBirthdayCommand struct {
	birthdays repository.BirthdayRepo
}

func (a *addBirthdayCommand) Name() string {
	return addBirthday
}

func (a *addBirthdayCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        addBirthday,
		Description: "Adds the birthday of a user",
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

	var response string
	if err != nil {
		log.Println(err.Error())
		response = err.Error()
	} else {
		err := a.birthdays.AddBirthday(birthdayUser)
		if err != nil {
			log.Println(err.Error())
			response = err.Error()
		}
	}

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

func (a *addBirthdayCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (repository.User, error) {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Or convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var errs error
	var birthdayUser repository.User

	if option, ok := optionMap[paramUser]; ok {
		usr := option.UserValue(s)
		birthdayUser.UserId = usr.ID
		birthdayUser.UserName = usr.Username
	} else {
		errs = errors.Join(errors.New("you need to specify the birthday user"))
	}

	if opt, ok := optionMap[paramBirthday]; ok {
		birthdayString := opt.StringValue()
		birthday, err := time.Parse(birthdayFormat, birthdayString)
		if err != nil {
			errs = errors.Join(fmt.Errorf("the birthday has to be in the format '%s'}", birthdayFormatReadable))

		}
		birthdayUser.Birthday = birthday
	} else {
		errs = errors.Join(fmt.Errorf("you need to specify the birthday date. Format '%s'", birthdayFormatReadable))
	}

	return birthdayUser, errs
}

func AddBirthday(repo repository.BirthdayRepo) BotCommand {
	cmd := new(addBirthdayCommand)
	cmd.birthdays = repo
	return cmd
}
