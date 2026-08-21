package birthday

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"

	"github.com/bwmarrin/discordgo"
)

const (
	setBirthdayRoleCommandName = "set-birthday-role"
	paramRole                  = "role"
)

type setBirthdayRoleCommand struct {
	roles *RoleRepository
}

func SetBirthdayRole(roles *RoleRepository) bot.Command {
	return &setBirthdayRoleCommand{roles: roles}
}

func (c *setBirthdayRoleCommand) Name() string {
	return setBirthdayRoleCommandName
}

func (c *setBirthdayRoleCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(discordgo.PermissionManageRoles)

	return &discordgo.ApplicationCommand{
		Name:                     setBirthdayRoleCommandName,
		Description:              "Sets the role that is given to users on their birthday",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        paramRole,
				Description: "Role to assign to users on their birthday",
				Required:    true,
			},
		},
	}
}

func (c *setBirthdayRoleCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	role, err := c.validateUserInput(s, i)

	var response string
	if err != nil {
		slog.Warn("invalid set-birthday-role input", "error", err)
		response = err.Error()
	} else if err := c.releasePreviousRole(ctx, s, i.GuildID, role.ID); err != nil {
		slog.Warn("error releasing previous birthday role", "error", err)
		response = "something went horribly wrong D:"
	} else if err := c.roles.SetBirthdayRoleId(ctx, i.GuildID, role.ID, true); err != nil {
		slog.Warn("error setting birthday role", "error", err)
		response = "something went horribly wrong D:"
	} else {
		response = fmt.Sprintf("birthday role set to %s", role.Mention())
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

func (c *setBirthdayRoleCommand) validateUserInput(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.Role, error) {
	optionMap := bot.OptionMap(i.ApplicationCommandData().Options)

	option, ok := optionMap[paramRole]
	if !ok {
		return nil, errors.New("you need to specify the birthday role")
	}

	role := option.RoleValue(s, i.GuildID)

	// The bot grants this role automatically, without re-checking who's
	// running the command each time, so it must never be a role that could
	// hand out more than a birthday: not @everyone, not an integration role
	// the bot can't actually (un)assign, and nothing carrying Administrator
	// or role/permission-management rights that a member could otherwise
	// only grant themselves via the normal Discord role hierarchy checks.
	if role.ID == i.GuildID {
		return nil, errors.New("the birthday role can't be @everyone")
	}
	if role.Managed {
		return nil, errors.New("that role is managed by an integration and can't be assigned by the bot")
	}
	const dangerousPermissions = discordgo.PermissionAdministrator |
		discordgo.PermissionManageRoles |
		discordgo.PermissionManageGuild |
		discordgo.PermissionManageWebhooks
	if role.Permissions&dangerousPermissions != 0 {
		return nil, errors.New("that role has admin-level permissions, pick a less privileged role for the birthday role")
	}

	return role, nil
}

// releasePreviousRole strips the previously configured birthday role from
// everyone we know still wears it, so switching roles doesn't leave the old
// one stuck on last cycle's birthday users.
func (c *setBirthdayRoleCommand) releasePreviousRole(ctx context.Context, s *discordgo.Session, guildId, newRoleId string) error {
	previous, err := c.roles.GetBirthdayRole(ctx, guildId)
	if err != nil {
		return err
	}
	if previous.RoleId == "" || previous.RoleId == newRoleId {
		return nil
	}

	holders, err := c.roles.GetHolders(ctx, guildId)
	if err != nil {
		return err
	}

	for _, holder := range holders {
		if err := s.GuildMemberRoleRemove(guildId, holder.UserId, previous.RoleId); err != nil {
			slog.Warn("error removing previous birthday role from user", "guild_id", guildId, "user_id", holder.UserId, "error", err)
		}
	}

	return c.roles.ClearHolders(ctx, guildId)
}
