package tests

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/shreyafeo/content-control-plane/internal/domain"
	"github.com/shreyafeo/content-control-plane/internal/repository"
)

// stubStore: in-memory fake for sync flow tests (counts upserts, tracks sync run id).
type stubStore struct {
	runID       uuid.UUID
	upsertCalls int
}

func (s *stubStore) ListPodcasts(ctx context.Context) ([]domain.Podcast, error) {
	return nil, nil
}

func (s *stubStore) GetPodcastByID(ctx context.Context, id uuid.UUID) (domain.Podcast, error) {
	return domain.Podcast{}, repository.ErrNotFound
}

func (s *stubStore) UpsertPodcast(ctx context.Context, p domain.Podcast) (uuid.UUID, error) {
	s.upsertCalls++
	return uuid.New(), nil
}

func (s *stubStore) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error {
	return nil
}

func (s *stubStore) CreateSyncRun(ctx context.Context, subject string) (domain.SyncRun, error) {
	s.runID = uuid.New()
	return domain.SyncRun{ID: s.runID, Subject: subject, Status: "running"}, nil
}

func (s *stubStore) CompleteSyncRun(ctx context.Context, id uuid.UUID, status string, count int, errMsg *string) error {
	return nil
}

func (s *stubStore) GetSyncRun(ctx context.Context, id uuid.UUID) (domain.SyncRun, error) {
	if id != s.runID {
		return domain.SyncRun{}, repository.ErrNotFound
	}
	return domain.SyncRun{
		ID:               s.runID,
		Status:           "success",
		RecordsProcessed: s.upsertCalls,
	}, nil
}

func (s *stubStore) InsertAudit(ctx context.Context, action, entityID string, metadata json.RawMessage) error {
	return nil
}

func (s *stubStore) ListAuditLogs(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	return nil, nil
}

// noopStore: Sync returns before touching the repo when query is empty; methods unused in those tests.
type noopStore struct{}

func (noopStore) ListPodcasts(ctx context.Context) ([]domain.Podcast, error) {
	return nil, nil
}
func (noopStore) GetPodcastByID(ctx context.Context, id uuid.UUID) (domain.Podcast, error) {
	return domain.Podcast{}, nil
}
func (noopStore) UpsertPodcast(ctx context.Context, p domain.Podcast) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (noopStore) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error { return nil }
func (noopStore) CreateSyncRun(ctx context.Context, subject string) (domain.SyncRun, error) {
	return domain.SyncRun{}, nil
}
func (noopStore) CompleteSyncRun(ctx context.Context, id uuid.UUID, status string, count int, errMsg *string) error {
	return nil
}
func (noopStore) GetSyncRun(ctx context.Context, id uuid.UUID) (domain.SyncRun, error) {
	return domain.SyncRun{}, nil
}
func (noopStore) InsertAudit(ctx context.Context, action, entityID string, metadata json.RawMessage) error {
	return nil
}
func (noopStore) ListAuditLogs(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	return nil, nil
}
