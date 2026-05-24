package ai

import (
	"context"

	"github.com/shreyafeo/content-control-plane/internal/domain"
)

// Input is the non-secret context passed to the model (also stored in ai_proposals.context).
type Input struct {
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Categories  []string `json:"categories"`
	FeedURL     string   `json:"feed_url"`
	SourceID    string   `json:"source_id"`
	ArtworkURL  string   `json:"artwork_url,omitempty"`
	TrackCount  *int     `json:"track_count,omitempty"`
}

// Result is a structured model response plus bookkeeping for audit.
type Result struct {
	Suggestion domain.MetadataSuggestion
	LatencyMS  int
	Model      string
	Provider   string
}

// Suggester produces validated structured metadata suggestions (never writes to the DB).
type Suggester interface {
	SuggestMetadata(ctx context.Context, in Input) (Result, error)
}
