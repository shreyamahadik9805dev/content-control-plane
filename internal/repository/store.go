package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/shreyafeo/content-control-plane/internal/domain"
)

// Store is the persistence boundary used by the service layer (mock or swap in tests).
type Store interface {
	ListPodcasts(ctx context.Context) ([]domain.Podcast, error)
	GetPodcastByID(ctx context.Context, id uuid.UUID) (domain.Podcast, error)
	UpsertPodcast(ctx context.Context, p domain.Podcast) (uuid.UUID, error)
	SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error
	CreateSyncRun(ctx context.Context, subject string) (domain.SyncRun, error)
	CompleteSyncRun(ctx context.Context, id uuid.UUID, status string, count int, errMsg *string) error
	GetSyncRun(ctx context.Context, id uuid.UUID) (domain.SyncRun, error)
	InsertAudit(ctx context.Context, action, entityID string, metadata json.RawMessage) error
	ListAuditLogs(ctx context.Context, limit int) ([]domain.AuditLog, error)

	InsertProposal(ctx context.Context, p domain.AIProposal) (uuid.UUID, error)
	GetProposalByID(ctx context.Context, id uuid.UUID) (domain.AIProposal, error)
	ListProposalsForPodcast(ctx context.Context, podcastID uuid.UUID, status string) ([]domain.AIProposal, error)
	SetProposalStatus(ctx context.Context, id uuid.UUID, status string, resolvedAt time.Time) error
	UpdatePodcastEnrichment(ctx context.Context, id uuid.UUID, summary string, operatorTags []string) error
}

var _ Store = (*Postgres)(nil)
