package features

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"github.com/go-co-op/gocron"
	"main/src/commands"
	"main/src/repository"
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

type Birthday struct {
	session      *discordgo.Session
	birthdayRepo repository.BirthdayRepo
	birthdayRole map[string]*discordgo.Role
}

func BirthdayRole(session *discordgo.Session) BotFeature {
	bd := new(Birthday)
	bd.session = session
	bd.birthdayRole = make(map[string]*discordgo.Role)
	bd.birthdayRepo = repository.NewBirthdayRepo()
	return bd
}

func (b Birthday) Init(session *discordgo.Session) error {
	guilds := session.State.Guilds
	if len(guilds) <= 0 {
		return errors.New("no guilds to serve")
	}

	for _, guild := range guilds {
		// init birthday role
		birthdayRole, err := session.GuildRoleCreate(guild.ID, &birthdayRole)
		if err != nil {
			return err
		}
		b.birthdayRole[guild.ID] = birthdayRole
	}

	err := b.scheduleBirthdayCheck()

	return err
}

func (b Birthday) Name() string {
	return featureBirthday
}

func (b Birthday) Commands() []commands.BotCommand {
	return []commands.BotCommand{
		commands.AddBirthday(b.birthdayRepo),
		commands.RemoveBirthday(b.birthdayRepo),
	}
}

func (b Birthday) scheduleBirthdayCheck() error {
	s := gocron.NewScheduler(time.UTC)

	_, err := s.Every(1).Day().At("0:30").Do(b.birthdayCheck)
	if err != nil {
		return err
	}

	s.StartAsync()
	return nil
}

func (b Birthday) birthdayCheck() error {
	guilds := b.session.State.Guilds

	for _, guild := range guilds {
		go func(guildId string) {
			err := b.ResetBirthdayRole(guildId)
			if err != nil {
				log.Printf(log.WARN, "Guild-%s: error resetting birthday role: %v \n", guildId, err.Error())
			}

			birthdayUsers, err := b.birthdayRepo.GetBirthdayUsers(time.Now())
			if err != nil {
				log.Printf(log.WARN, "Guild-%s: error getting birthday users: %v \n", guildId, err.Error())
			}

			birthdayRole := b.birthdayRole[guildId]
			for _, user := range birthdayUsers {
				err := b.session.GuildMemberRoleAdd(guildId, user.UserId, birthdayRole.ID)
				if err != nil {
					log.Printf(log.WARN, "Guild-%s: error setting birthday role for user-%s: %v \n", guildId, user.UserId, err.Error())
				}
			}
		}(guild.ID)
	}

	return nil
}

func (b Birthday) ResetBirthdayRole(guildId string) error {
	if role, ok := b.birthdayRole[guildId]; ok {
		err := b.session.GuildRoleDelete(guildId, role.ID)
		if err != nil {
			return err
		}
	}

	birthdayRole, err := b.session.GuildRoleCreate(guildId, &birthdayRole)
	if err != nil {
		return err
	}
	b.birthdayRole[guildId] = birthdayRole

	return nil
}
