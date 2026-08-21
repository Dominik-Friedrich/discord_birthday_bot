// Package youtube wraps github.com/wader/goutubedl (which execs yt-dlp from
// PATH) to resolve a query or URL to track metadata.
package youtube

import (
	"context"
	"fmt"
	"net/url"

	"github.com/wader/goutubedl"
)

// Resolve looks up metadata for a query, which may be a YouTube URL or a
// free-text search term.
func Resolve(ctx context.Context, query string) (goutubedl.Result, error) {
	target := query
	if !isURL(query) {
		videoURL, err := searchTopResult(ctx, query)
		if err != nil {
			return goutubedl.Result{}, err
		}
		target = videoURL
	}

	// TypeSingle rejects bare playlist URLs.
	return goutubedl.New(ctx, target, goutubedl.Options{Type: goutubedl.TypeSingle})
}

// searchTopResult finds the top hit for query and returns its own URL.
// yt-dlp represents search results as a playlist, which TypeSingle rejects,
// so search resolves in two steps: find the hit here (TypeAny), then let the
// caller re-resolve it as a normal single video.
func searchTopResult(ctx context.Context, query string) (string, error) {
	result, err := goutubedl.New(ctx, "ytsearch:"+query, goutubedl.Options{Type: goutubedl.TypeAny})
	if err != nil {
		return "", err
	}
	if len(result.Info.Entries) == 0 {
		return "", fmt.Errorf("no results for %q", query)
	}

	top := result.Info.Entries[0]
	if top.WebpageURL != "" {
		return top.WebpageURL, nil
	}
	if top.ID != "" {
		return top.ID, nil
	}
	return "", fmt.Errorf("no results for %q", query)
}

func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}
