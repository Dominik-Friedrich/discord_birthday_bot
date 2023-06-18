package features

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"github.com/go-co-op/gocron"
	"main/src/bot"
	"main/src/commands"
	"main/src/database"
	"main/src/repository/birthday"
	"time"
)

const (
	featureBirthday = "featureBirthday"
)

var (
	birthdayRoleName = "Geburtstagskind"
	// decimal for hex FFFB00, bright yellow color
	birthdayRoleColor       = 16775936
	birthdayRoleHoist       = true
	birthdayRolePermissions = int64(discordgo.PermissionViewChannel)
	birthdayRoleMentionable = true
	birthdayRole            = discordgo.RoleParams{
		Name:        birthdayRoleName,
		Color:       &birthdayRoleColor,
		Hoist:       &birthdayRoleHoist,
		Permissions: &birthdayRolePermissions,
		Mentionable: &birthdayRoleMentionable,
	}
)

// TODO: remember which user has birthday role instead of deleting and recreating the entire role?

type Birthday struct {
	session      *bot.Session
	birthdayRepo birthday.Repository
}

func BirthdayRole(connection *database.Connection) bot.Feature {
	b := new(Birthday)
	b.birthdayRepo = birthday.NewRepository(connection)
	return b
}

func (b Birthday) Init(session *bot.Session) error {
	b.session = session

	err := b.scheduleBirthdayCheck()

	return err
}

func (b Birthday) Name() string {
	return featureBirthday
}

func (b Birthday) Commands() []bot.Command {
	return []bot.Command{
		commands.AddBirthday(b.birthdayRepo),
		commands.RemoveBirthday(b.birthdayRepo),
	}
}

func (b Birthday) scheduleBirthdayCheck() error {
	s := gocron.NewScheduler(time.Local)

	_, err := s.Every(1).Day().At("0:30").StartImmediately().Do(b.birthdayCheckGuilds)
	if err != nil {
		return err
	}

	s.StartAsync()
	return nil
}

func (b Birthday) birthdayCheckGuilds() error {
	guilds := b.session.State.Guilds

	for _, guild := range guilds {
		go b.asyncBirthdayCheckGuild(guild.ID)
	}

	return nil
}

func (b Birthday) asyncBirthdayCheckGuild(guildId string) {
	err := b.resetBirthdayRole(guildId)
	if err != nil {
		log.Printf(log.WARN, "Guild-%s: error resetting birthday role: %v \n", guildId, err.Error())
	}

	birthdayUsers, err := b.birthdayRepo.GetBirthdayUsers(time.Now())
	if err != nil {
		log.Printf(log.WARN, "Guild-%s: error getting birthday users: %v \n", guildId, err.Error())
	}
	if len(birthdayUsers) <= 0 {
		return
	}

	log.Println("Guild-%s: birthday users found:")
	log.PrettyPrint(log.INFO, birthdayUsers)

	roleId, err := b.birthdayRepo.GetBirthdayRoleId(guildId)
	if err != nil {
		log.Printf(log.WARN, "Guild-%s: error getting birthday role id: %v \n", guildId, err.Error())
	}

	for _, user := range birthdayUsers {
		err := b.session.GuildMemberRoleAdd(guildId, user.UserId, roleId)
		if err != nil {
			log.Printf(log.WARN, "Guild-%s: error setting birthday role for user-%s: %v \n", guildId, user.UserId, err.Error())
		}
	}
}

func (b Birthday) resetBirthdayRole(guildId string) error {
	roleId, err := b.birthdayRepo.GetBirthdayRoleId(guildId)
	if err != nil {
		return fmt.Errorf("error getting birthday role id: %v", err.Error())
	}

	// only delete if it existed
	if roleId != "" {
		err := b.session.GuildRoleDelete(guildId, roleId)
		if err != nil {
			log.Printf(log.WARN, "Guild-%s error deleting birthday role: %v \n", guildId, err.Error())
		}
	}

	birthdayRole, err := b.session.GuildRoleCreate(guildId, &birthdayRole)
	if err != nil {
		return err
	}

	err = b.birthdayRepo.SetBirthdayRoleId(guildId, birthdayRole.ID)
	if err != nil {
		return err
	}

	err = b.setBirthdayAsHighAsPossible(guildId)

	return err
}

func (b Birthday) setBirthdayAsHighAsPossible(guildId string) error {
	roles, err := b.session.GuildRoles(guildId)
	if err != nil {
		return fmt.Errorf("error getting guild roles: %v", err.Error())
	}

	botUser, err := b.session.GuildMember(guildId, b.session.ApplicationId)
	if err != nil {
		return fmt.Errorf("error getting bot user: %v", err.Error())
	}

	birthdayRoleId, err := b.birthdayRepo.GetBirthdayRoleId(guildId)
	if err != nil {
		return fmt.Errorf("error getting birthday role id: %v", err.Error())
	}
	if birthdayRoleId == "" {
		return errors.New("birthday role does not exist")
	}

	// this should be fine as servers probably don't have too many roles
	var botRolePosition int
	var birthdayRoleIndex int
	for i, role := range roles {
		if role.ID == birthdayRoleId {
			birthdayRoleIndex = i
		}

		for _, botRoleId := range botUser.Roles {
			if botRoleId == role.ID && role.Managed {
				botRolePosition = role.Position
			}
		}
	}

	roles[birthdayRoleIndex].Position = botRolePosition

	_, err = b.session.GuildRoleReorder(guildId, roles)

	return err
}
