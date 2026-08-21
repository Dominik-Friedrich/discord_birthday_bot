package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	// The runtime image has no OS timezone database, so embed Go's copy
	// directly into the binary. Without this, birthday.Birthday's
	// gocron.WithLocation(time.Local) - and any TZ env var an operator
	// sets - would silently resolve to UTC.
	_ "time/tzdata"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/birthday"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/complaint"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/config"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/database"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/player"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/retry"
	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/user"
)

// databaseRetry paces reconnect attempts if the database isn't up yet when
// the bot starts (e.g. compose brought up the bot before postgres finished
// initializing). 20 attempts with a 30s cap gives postgres roughly 8 minutes
// to become reachable before this is treated as a real failure instead of a
// slow start.
var databaseRetry = retry.Config{
	MaxAttempts: 20,
	BaseDelay:   time.Second,
	MaxDelay:    30 * time.Second,
}

func main() {
	cfg := config.Load()
	setupLogging(cfg.LogLevel)

	// run() reports every startup failure as a plain error; this is the
	// only place in the program allowed to panic, so a human (or the
	// container orchestrator, via the non-zero exit) always has exactly one
	// place to look for why the bot went down.
	if err := run(cfg); err != nil {
		slog.Error("fatal error", "error", err)
		panic(err)
	}
}

func run(cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	db, err := connectDatabase(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	// userRepo must be constructed (and migrated) before complaintRepo, since
	// Complaint/Reply carry foreign keys into the users table.
	userRepo, err := user.NewRepository(db)
	if err != nil {
		return fmt.Errorf("initializing user repository: %w", err)
	}

	roleRepo, err := birthday.NewRoleRepository(db)
	if err != nil {
		return fmt.Errorf("initializing birthday role repository: %w", err)
	}

	complaintRepo, err := complaint.NewRepository(db, userRepo)
	if err != nil {
		return fmt.Errorf("initializing complaint repository: %w", err)
	}

	birthdayBot, err := bot.New(cfg.Discord.Token, cfg.Discord.ApplicationID)
	if err != nil {
		return fmt.Errorf("initializing bot: %w", err)
	}

	birthdayBot.RegisterFeature(birthday.New(userRepo, roleRepo))
	birthdayBot.RegisterFeature(complaint.New(complaintRepo))
	birthdayBot.RegisterFeature(player.New(cfg.Player.MaxMediaDuration))

	return birthdayBot.Run(ctx)
}

func connectDatabase(ctx context.Context, cfg database.Config) (*database.Connection, error) {
	var db *database.Connection

	err := retry.Do(ctx, "database connection", databaseRetry, func() error {
		var connErr error
		db, connErr = database.NewConnection(cfg)
		return connErr
	})

	return db, err
}

func setupLogging(level string) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})))
}
