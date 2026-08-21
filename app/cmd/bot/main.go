package main

import (
	"log/slog"
	"os"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/birthday"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/complaint"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/config"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/database"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/user"
)

func main() {
	cfg := config.Load()
	setupLogging(cfg.LogLevel)

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		slog.Error("error initializing db connection", "error", err)
		panic(err)
	}

	// userRepo must be constructed (and migrated) before complaintRepo, since
	// Complaint/Reply carry foreign keys into the users table.
	userRepo := user.NewRepository(db)
	roleRepo := birthday.NewRoleRepository(db)
	complaintRepo := complaint.NewRepository(db, userRepo)

	birthdayBot := bot.New(cfg.Discord.Token, cfg.Discord.ApplicationID)
	birthdayBot.RegisterFeature(birthday.New(userRepo, roleRepo))
	birthdayBot.RegisterFeature(complaint.New(complaintRepo))

	birthdayBot.Run()
}

func setupLogging(level string) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})))
}
