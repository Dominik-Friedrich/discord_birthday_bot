// Package retry provides a small exponential-backoff-with-jitter helper for
// surviving transient startup failures (e.g. the database not being ready
// yet) without crash-looping the whole process.
package retry

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"
)

// Config controls how Do paces its retries.
type Config struct {
	// MaxAttempts caps how many times fn is called. Zero means retry until
	// ctx is cancelled.
	MaxAttempts int
	// BaseDelay is the wait before the first retry; it doubles after every
	// subsequent failure, capped at MaxDelay.
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// Do calls fn until it succeeds, ctx is cancelled, or MaxAttempts is
// exhausted, sleeping with exponential backoff and full jitter between
// attempts (see https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)
// so retries don't pile up in lockstep.
func Do(ctx context.Context, operation string, cfg Config, fn func() error) error {
	delay := cfg.BaseDelay
	if delay <= 0 {
		delay = time.Second
	}

	for attempt := 1; ; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		if cfg.MaxAttempts > 0 && attempt >= cfg.MaxAttempts {
			return fmt.Errorf("%s: giving up after %d attempts: %w", operation, attempt, err)
		}

		wait := time.Duration(rand.Int64N(int64(delay)))
		slog.Warn("retrying after failure", "operation", operation, "attempt", attempt, "wait", wait, "error", err)

		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", operation, ctx.Err())
		case <-time.After(wait):
		}

		delay *= 2
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}
}
