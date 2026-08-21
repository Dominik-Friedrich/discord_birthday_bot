// Package birthday implements the birthday-role feature: users can be given
// a birthday, and once a day the bot grants a special role to anyone whose
// birthday it is.
package birthday

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/user"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"
)

const featureBirthday = "featureBirthday"

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

// UserRepository is the subset of user.Repository the birthday feature needs.
type UserRepository interface {
	UpsertUser(ctx context.Context, u *user.User) error
	RemoveBirthday(ctx context.Context, u user.User) error
	GetUsers(ctx context.Context, birthday time.Time) ([]user.User, error)
}

type Birthday struct {
	session            *bot.Session
	users              UserRepository
	roles              *RoleRepository
	birthdayAddedEvent chan user.User
}

func New(users UserRepository, roles *RoleRepository) bot.Feature {
	b := new(Birthday)

	b.birthdayAddedEvent = make(chan user.User)
	b.users = users
	b.roles = roles

	return b
}

func (b *Birthday) Init(session *bot.Session) error {
	b.session = session

	err := b.scheduleBirthdayCheck()
	go b.watchBirthdayAddedEvents()

	return err
}

func (b *Birthday) Name() string {
	return featureBirthday
}

func (b *Birthday) Commands() []bot.Command {
	return []bot.Command{
		AddBirthday(b.users, b.birthdayAddedEvent),
		RemoveBirthday(b.users),
		SetBirthdayRole(b.roles),
	}
}

// watchBirthdayAddedEvents reacts to a birthday being added for today's date
// by immediately re-checking that guild, instead of waiting for the next
// scheduled run.
func (b *Birthday) watchBirthdayAddedEvents() {
	for birthdayUser := range b.birthdayAddedEvent {
		slog.Debug("birthday added event", "guild_id", birthdayUser.GuildId, "user_id", birthdayUser.UserId)

		today := time.Now()
		if today.Day() == birthdayUser.Birthday.Day() && today.Month() == birthdayUser.Birthday.Month() {
			go b.asyncBirthdayCheckGuild(context.Background(), birthdayUser.GuildId)
		}
	}
}

func (b *Birthday) scheduleBirthdayCheck() error {
	s, err := gocron.NewScheduler(gocron.WithLocation(time.Local))
	if err != nil {
		return err
	}

	_, err = s.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(0, 30, 0))),
		gocron.NewTask(b.birthdayCheckGuilds),
		gocron.WithStartAt(gocron.WithStartImmediately()),
	)
	if err != nil {
		return err
	}

	s.Start()
	return nil
}

func (b *Birthday) birthdayCheckGuilds() {
	ctx := context.Background()
	for _, guild := range b.session.State.Guilds {
		go b.asyncBirthdayCheckGuild(ctx, guild.ID)
	}
}

func (b *Birthday) asyncBirthdayCheckGuild(ctx context.Context, guildId string) {
	birthdayUsers, err := b.users.GetUsers(ctx, time.Now())
	if err != nil {
		slog.Warn("error getting birthday users", "guild_id", guildId, "error", err)
	}
	if len(birthdayUsers) <= 0 {
		return
	}

	role, err := b.ensureBirthdayRole(ctx, guildId)
	if err != nil {
		slog.Warn("error ensuring birthday role", "guild_id", guildId, "error", err)
		return
	}

	if err := b.releaseBirthdayRoleHolders(ctx, guildId, role.RoleId); err != nil {
		slog.Warn("error releasing previous birthday role holders", "guild_id", guildId, "error", err)
	}

	slog.Info("birthday users found", "guild_id", guildId, "count", len(birthdayUsers))

	for _, u := range birthdayUsers {
		if err := b.session.GuildMemberRoleAdd(guildId, u.UserId, role.RoleId); err != nil {
			slog.Warn("error setting birthday role for user", "guild_id", guildId, "user_id", u.UserId, "error", err)
			continue
		}
		if err := b.roles.AddHolder(ctx, guildId, u.UserId); err != nil {
			slog.Warn("error tracking birthday role holder", "guild_id", guildId, "user_id", u.UserId, "error", err)
		}
	}
}

// ensureBirthdayRole returns the role to use for today's birthday users,
// auto-creating and positioning a default role the first time a guild is
// checked and no role (auto or admin-picked via /set-birthday-role) exists
// yet. An already-configured role, custom or not, is reused as-is.
func (b *Birthday) ensureBirthdayRole(ctx context.Context, guildId string) (Role, error) {
	role, err := b.roles.GetBirthdayRole(ctx, guildId)
	if err != nil {
		return Role{}, fmt.Errorf("error getting birthday role: %w", err)
	}
	if role.RoleId != "" {
		return role, nil
	}

	newRole, err := b.session.GuildRoleCreate(guildId, &birthdayRole)
	if err != nil {
		return Role{}, err
	}

	role = Role{GuildId: guildId, RoleId: newRole.ID, Custom: false}
	if err := b.roles.SetBirthdayRoleId(ctx, guildId, role.RoleId, role.Custom); err != nil {
		return Role{}, err
	}

	if err := b.setBirthdayAsHighAsPossible(guildId, role.RoleId); err != nil {
		return Role{}, err
	}

	return role, nil
}

// releaseBirthdayRoleHolders removes the birthday role from everyone it was
// previously handed out to, using our own record of who holds it instead of
// deleting/recreating the role (which would destroy an admin-picked role)
// or listing guild members (which needs the privileged members intent).
func (b *Birthday) releaseBirthdayRoleHolders(ctx context.Context, guildId, roleId string) error {
	holders, err := b.roles.GetHolders(ctx, guildId)
	if err != nil {
		return fmt.Errorf("error getting birthday role holders: %w", err)
	}

	for _, holder := range holders {
		if err := b.session.GuildMemberRoleRemove(guildId, holder.UserId, roleId); err != nil {
			slog.Warn("error removing birthday role from user", "guild_id", guildId, "user_id", holder.UserId, "error", err)
		}
	}

	return b.roles.ClearHolders(ctx, guildId)
}

func (b *Birthday) setBirthdayAsHighAsPossible(guildId, birthdayRoleId string) error {
	roles, err := b.session.GuildRoles(guildId)
	if err != nil {
		return fmt.Errorf("error getting guild roles: %w", err)
	}

	botUser, err := b.session.GuildMember(guildId, b.session.ApplicationId)
	if err != nil {
		return fmt.Errorf("error getting bot user: %w", err)
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
