package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Podcast struct {
	ID          uuid.UUID `json:"id"`
	SourceID    string    `json:"source_id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Categories  []string  `json:"categories"`
	FeedURL     string    `json:"feed_url"`
	ArtworkURL  string    `json:"artwork_url"`
	TrackCount  *int      `json:"track_count,omitempty"`
	Pinned      bool      `json:"pinned"`
	Featured    bool      `json:"featured"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
