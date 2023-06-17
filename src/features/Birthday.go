package features

import (
	"github.com/bwmarrin/discordgo"
	"main/src/commands"
	"main/src/repository"
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
	birthdayRepo repository.BirthdayRepo
	birthdayRole map[string]*discordgo.Role
}

func BirthdayRole() BotFeature {
	bd := new(Birthday)
	bd.birthdayRole = make(map[string]*discordgo.Role)
	bd.birthdayRepo = repository.NewBirthdayRepo()
	return bd
}

func (b Birthday) Init(session *discordgo.Session) error {
	for _, guild := range session.State.Guilds {
		// init birthday role
		birthdayRole, err := session.GuildRoleCreate(guild.ID, &birthdayRole)
		if err != nil {
			return err
		}
		b.birthdayRole[guild.ID] = birthdayRole
	}

	return nil
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
