package youtube

import (
	"context"
	"testing"
	"time"
)

func Test_isURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"full url", "https://www.youtube.com/watch?v=n5lt-y8RcVc", true},
		{"shorts url", "https://www.youtube.com/shorts/Pjj_tBNh0-I", true},
		{"bare search term", "never gonna give you up", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isURL(tt.in); got != tt.want {
				t.Errorf("isURL(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestResolve_URL hits the real yt-dlp binary and network. Requires yt-dlp
// on PATH.
func TestResolve_URL(t *testing.T) {
	const testQuery = "https://www.youtube.com/watch?v=EIyixC9NsLI"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := Resolve(ctx, testQuery)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.Info.ID != "EIyixC9NsLI" {
		t.Errorf("Info.ID = %q, want %q", result.Info.ID, "EIyixC9NsLI")
	}
	if result.Info.Duration <= 0 {
		t.Errorf("Info.Duration = %v, want > 0", result.Info.Duration)
	}
}

// TestResolve_Search covers the ytsearch: prefixing path for free-text
// queries.
func TestResolve_Search(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := Resolve(ctx, "rick astley never gonna give you up")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.Info.ID == "" {
		t.Error("Info.ID is empty, expected a resolved video")
	}
}
