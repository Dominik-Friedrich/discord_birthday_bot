// Package user holds the User entity shared by the birthday and complaint
// features (a Discord member has at most one User row per guild).
package user

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	gorm.Model
	GuildId  string `gorm:"index:idx_guildUser,unique"`
	UserId   string `gorm:"index:idx_guildUser,unique"`
	Username string
	Nickname *string
	Birthday time.Time
}

type Repository struct {
	db *database.Connection
}

func NewRepository(db *database.Connection) *Repository {
	if err := db.AutoMigrate(&User{}); err != nil {
		slog.Error("error migrating user schema", "error", err)
		panic(err)
	}

	return &Repository{db: db}
}

func (r *Repository) GetUsers(ctx context.Context, birthday time.Time) ([]User, error) {
	slog.Debug("getting users by birthday", "day", birthday.Day(), "month", birthday.Month())

	return gorm.G[User](r.db.DB).
		Where("EXTRACT(MONTH FROM birthday) = ? AND EXTRACT(DAY FROM birthday) = ?", birthday.Month(), birthday.Day()).
		Find(ctx)
}

func (r *Repository) UpsertUser(ctx context.Context, user *User) error {
	slog.Info("upserting user", "guild_id", user.GuildId, "user_id", user.UserId)

	return gorm.G[User](r.db.DB, clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "guild_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"birthday", "username", "nickname"}),
	}).Create(ctx, user)
}

func (r *Repository) GetOrCreateUser(ctx context.Context, user *User) error {
	existingUser, err := gorm.G[User](r.db.DB).Where("user_id = ? AND guild_id = ?", user.UserId, user.GuildId).First(ctx)
	if err == nil {
		*user = existingUser
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := gorm.G[User](r.db.DB).Create(ctx, user); err != nil {
			return err
		}
		slog.Info("created user", "guild_id", user.GuildId, "user_id", user.UserId)
		return nil
	}

	return err
}

func (r *Repository) RemoveBirthday(ctx context.Context, user User) error {
	slog.Info("removing user", "guild_id", user.GuildId, "user_id", user.UserId)

	_, err := gorm.G[User](r.db.Unscoped()).Where(&User{GuildId: user.GuildId, UserId: user.UserId}).Delete(ctx)
	return err
}
