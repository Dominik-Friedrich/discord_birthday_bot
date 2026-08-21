package player

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wader/goutubedl"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/player/youtube"
)

// resolver looks up a track and rejects it up front if it's over
// maxDuration, instead of failing later when its turn to play comes up.
type resolver struct {
	maxDuration time.Duration
}

func newResolver(maxDuration time.Duration) *resolver {
	return &resolver{maxDuration: maxDuration}
}

func (r *resolver) resolve(ctx context.Context, query string) (goutubedl.Result, error) {
	result, err := youtube.Resolve(ctx, query)
	if err != nil {
		return goutubedl.Result{}, fmt.Errorf("resolving %q: %w", query, err)
	}

	if r.maxDuration > 0 {
		duration := time.Duration(result.Info.Duration * float64(time.Second))
		if duration > r.maxDuration {
			slog.Debug("rejecting track, too long", "title", result.Info.Title, "duration", duration, "max_duration", r.maxDuration)
			return goutubedl.Result{}, fmt.Errorf("video is too long (%s, max %s)", duration.Round(time.Second), r.maxDuration)
		}
	}

	return result, nil
}
