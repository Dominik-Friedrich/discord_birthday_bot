package commands

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
	"main/src/repository/birthday"
	"time"
)

const (
	addBirthday            = "add-birthday"
	paramUser              = "user"
	paramBirthday          = "birthday"
	birthdayFormat         = "02/01"
	birthdayFormatReadable = "DD/MM"
)

type addBirthdayCommand struct {
	birthdays birthday.Repository
}

func AddBirthday(repo birthday.Repository) bot.Command {
	cmd := new(addBirthdayCommand)
	cmd.birthdays = repo
	return cmd
}

func (a *addBirthdayCommand) Name() string {
	return addBirthday
}

func (a *addBirthdayCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionManageRoles)

	return &discordgo.ApplicationCommand{
		Name:                     addBirthday,
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
		log.Println(err.Error())
		response = err.Error()
	} else {
		err := a.birthdays.UpsertBirthday(birthdayUser)
		if err != nil {
			log.Println(log.WARN, err.Error())
			response = "something went horribly wrong D:"
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

func (a *addBirthdayCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (birthday.User, error) {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Or convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var errs error
	var birthdayUser birthday.User

	if option, ok := optionMap[paramUser]; ok {
		usr := option.UserValue(s)
		birthdayUser.UserId = usr.ID
		birthdayUser.GuildId = i.GuildID
		birthdayUser.UserName = usr.Username
	} else {
		errs = errors.Join(errors.New("you need to specify the birthday user"))
	}

	if opt, ok := optionMap[paramBirthday]; ok {
		birthdayString := opt.StringValue()
		birthdayDate, err := time.Parse(birthdayFormat, birthdayString)
		if err != nil {
			log.PrettyPrint(log.INFO, birthdayString)
			errs = errors.Join(fmt.Errorf("the birthday has to be in the format '%s'}", birthdayFormatReadable))

		}
		birthdayUser.Birthday = birthdayDate
	} else {
		errs = errors.Join(fmt.Errorf("you need to specify the birthday date. Format '%s'", birthdayFormatReadable))
	}

	return birthdayUser, errs
}
