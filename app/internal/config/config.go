package config

import (
	"time"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/database"

	"github.com/spf13/viper"
)

type Discord struct {
	Token         string
	ApplicationID string
}

type Player struct {
	// MaxMediaDuration rejects a track at queue time if it's longer than
	// this.
	MaxMediaDuration time.Duration
}

type Config struct {
	Database database.Config
	Discord  Discord
	Player   Player
	// LogLevel is one of debug, info, warn, error.
	LogLevel string
}

// Load binds the bot's environment variables (prefixed BOT_) and returns the
// resulting configuration.
func Load() Config {
	viper.SetEnvPrefix("bot")

	_ = viper.BindEnv("database_type")
	_ = viper.BindEnv("database_host")
	_ = viper.BindEnv("database_port")
	_ = viper.BindEnv("database_user")
	_ = viper.BindEnv("database_password")
	_ = viper.BindEnv("database_database")

	_ = viper.BindEnv("discord_token")
	_ = viper.BindEnv("discord_application_id")

	_ = viper.BindEnv("log_level")
	viper.SetDefault("log_level", "info")

	_ = viper.BindEnv("player_max_media_duration")
	viper.SetDefault("player_max_media_duration", "10m")

	viper.AutomaticEnv()

	return Config{
		Database: database.Config{
			Type:     viper.GetString("database_type"),
			Host:     viper.GetString("database_host"),
			Port:     viper.GetString("database_port"),
			User:     viper.GetString("database_user"),
			Password: viper.GetString("database_password"),
			Database: viper.GetString("database_database"),
		},
		Discord: Discord{
			Token:         viper.GetString("discord_token"),
			ApplicationID: viper.GetString("discord_application_id"),
		},
		Player: Player{
			MaxMediaDuration: viper.GetDuration("player_max_media_duration"),
		},
		LogLevel: viper.GetString("log_level"),
	}
}
