package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Podcast struct {
	ID           uuid.UUID `json:"id"`
	SourceID     string    `json:"source_id"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	Categories   []string  `json:"categories"`
	FeedURL      string    `json:"feed_url"`
	ArtworkURL   string    `json:"artwork_url"`
	TrackCount   *int      `json:"track_count,omitempty"`
	Summary      string    `json:"summary"`
	OperatorTags []string  `json:"operator_tags"`
	Pinned       bool      `json:"pinned"`
	Featured     bool      `json:"featured"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SyncRun struct {
	ID               uuid.UUID  `json:"id"`
	Subject          string     `json:"subject"`
	Status           string     `json:"status"`
	RecordsProcessed int        `json:"records_processed"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

type AuditLog struct {
	ID        uuid.UUID       `json:"id"`
	Action    string          `json:"action"`
	EntityID  string          `json:"entity_id"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}

// AIProposal is a human-in-the-loop suggestion row; canonical catalog changes only after accept.
type AIProposal struct {
	ID         uuid.UUID       `json:"id"`
	PodcastID  uuid.UUID       `json:"podcast_id"`
	Status     string          `json:"status"`
	Kind       string          `json:"kind"`
	Payload    json.RawMessage `json:"payload"`
	Context    json.RawMessage `json:"context"`
	Model      string          `json:"model"`
	Provider   string          `json:"provider"`
	LatencyMS  *int            `json:"latency_ms,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	ResolvedAt *time.Time      `json:"resolved_at,omitempty"`
}

// MetadataSuggestion is the JSON shape we ask the model for (and validate before apply).
type MetadataSuggestion struct {
	Summary       string   `json:"summary"`
	OperatorTags  []string `json:"operator_tags"`
	Language      string   `json:"language"`
	Confidence    float64  `json:"confidence"`
}
