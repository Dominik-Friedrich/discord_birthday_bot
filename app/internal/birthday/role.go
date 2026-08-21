package birthday

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Role tracks which Discord role a guild uses to mark birthdays.
type Role struct {
	gorm.Model
	GuildId string `gorm:"unique"`
	RoleId  string
	// Custom marks a role an admin picked via /set-birthday-role, as opposed
	// to one the bot auto-created. The bot never deletes or repositions a
	// custom role.
	Custom bool
}

// RoleHolder tracks which users currently wear the birthday role, so it can
// be taken away from exactly the right people the next time the role is
// handed out, without deleting/recreating the role or needing the
// privileged guild members intent to enumerate current role members.
type RoleHolder struct {
	gorm.Model
	GuildId string `gorm:"index:idx_guildRoleHolder,unique"`
	UserId  string `gorm:"index:idx_guildRoleHolder,unique"`
}

type RoleRepository struct {
	db *database.Connection
}

func NewRoleRepository(db *database.Connection) *RoleRepository {
	if err := db.AutoMigrate(&Role{}, &RoleHolder{}); err != nil {
		slog.Error("error migrating birthday role schema", "error", err)
		panic(err)
	}

	return &RoleRepository{db: db}
}

func (r *RoleRepository) SetBirthdayRoleId(ctx context.Context, guildId, roleId string, custom bool) error {
	slog.Info("setting birthday role", "guild_id", guildId, "role_id", roleId, "custom", custom)

	return gorm.G[Role](r.db.DB, clause.OnConflict{
		Columns:   []clause.Column{{Name: "guild_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role_id", "custom"}),
	}).Create(ctx, &Role{GuildId: guildId, RoleId: roleId, Custom: custom})
}

func (r *RoleRepository) GetBirthdayRole(ctx context.Context, guildId string) (Role, error) {
	slog.Debug("getting birthday role", "guild_id", guildId)

	role, err := gorm.G[Role](r.db.DB).Where(&Role{GuildId: guildId}).First(ctx)

	// acceptable error, guild simply has no birthday role yet
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Role{}, nil
	}

	return role, err
}

// GetHolders returns the users currently tracked as wearing the birthday
// role in the given guild.
func (r *RoleRepository) GetHolders(ctx context.Context, guildId string) ([]RoleHolder, error) {
	return gorm.G[RoleHolder](r.db.DB).Where(&RoleHolder{GuildId: guildId}).Find(ctx)
}

func (r *RoleRepository) AddHolder(ctx context.Context, guildId, userId string) error {
	return gorm.G[RoleHolder](r.db.DB, clause.OnConflict{
		Columns:   []clause.Column{{Name: "guild_id"}, {Name: "user_id"}},
		DoNothing: true,
	}).Create(ctx, &RoleHolder{GuildId: guildId, UserId: userId})
}

// ClearHolders forgets every tracked holder for the guild. Call it once
// their roles have actually been removed on Discord's side.
func (r *RoleRepository) ClearHolders(ctx context.Context, guildId string) error {
	_, err := gorm.G[RoleHolder](r.db.Unscoped()).Where(&RoleHolder{GuildId: guildId}).Delete(ctx)
	return err
}
