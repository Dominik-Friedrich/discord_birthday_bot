package discord

import (
	"testing"
	"time"
)

// TestAudioPlayer_ControlCallsNeverBlock guards against the deadlock class
// where Stop/Pause/Unpause block because asyncPlayRoutine isn't listening
// for them at that instant (e.g. idle, or busy elsewhere) -- see the
// select-race finding this covers.
func TestAudioPlayer_ControlCallsNeverBlock(t *testing.T) {
	p, err := NewPlayer(make(chan PlayerContext))
	if err != nil {
		t.Fatalf("NewPlayer() error = %v", err)
	}

	for _, call := range []func(){p.Stop, p.Pause, p.Unpause} {
		done := make(chan struct{})
		go func() {
			call()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("call blocked for over a second")
		}
	}
}
