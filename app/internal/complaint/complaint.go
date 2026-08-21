// Package complaint implements the complaint feature: users can complain
// about something (optionally about another user), and the bot answers with
// a randomly chosen, guild-configurable reply.
package complaint

import (
	"context"
	"log/slog"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
)

const featureComplaint = "featureComplaint"

type Feature struct {
	session *bot.Session
	repo    *Repository
	replies *cache
}

func New(repo *Repository) bot.Feature {
	f := &Feature{
		repo:    repo,
		replies: new(cache),
	}

	if err := refreshCache(context.Background(), repo, f.replies); err != nil {
		slog.Warn("unable to load complaint replies", "error", err)
	}

	return f
}

func (f *Feature) Init(session *bot.Session) error {
	f.session = session

	return nil
}

func (f *Feature) Name() string {
	return featureComplaint
}

func (f *Feature) Commands() []bot.Command {
	return []bot.Command{
		Complain(f.repo, f.replies),
		AddComplainReply(f.repo, f.replies),
	}
}
