package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shreyafeo/content-control-plane/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres opens a pool, pings once, and bails early if the DB isn't reachable.
func NewPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

// ListPodcasts returns every row, newest updates first.
func (p *Postgres) ListPodcasts(ctx context.Context) ([]domain.Podcast, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, source_id, title, author, categories, feed_url, artwork_url, track_count,
		       pinned, featured, created_at, updated_at
		FROM podcasts
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPodcasts(rows)
}

// GetPodcastByID returns ErrNotFound when the UUID isn't in the table.
func (p *Postgres) GetPodcastByID(ctx context.Context, id uuid.UUID) (domain.Podcast, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT id, source_id, title, author, categories, feed_url, artwork_url, track_count,
		       pinned, featured, created_at, updated_at
		FROM podcasts WHERE id = $1
	`, id)
	pod, err := scanPodcastRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Podcast{}, ErrNotFound
	}
	return pod, err
}

// UpsertPodcast inserts or updates on source_id conflict and returns the row id.
func (p *Postgres) UpsertPodcast(ctx context.Context, pod domain.Podcast) (uuid.UUID, error) {
	cats, err := json.Marshal(pod.Categories)
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = p.pool.QueryRow(ctx, `
		INSERT INTO podcasts (source_id, title, author, categories, feed_url, artwork_url, track_count, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		ON CONFLICT (source_id) DO UPDATE SET
			title = EXCLUDED.title,
			author = EXCLUDED.author,
			categories = EXCLUDED.categories,
			feed_url = EXCLUDED.feed_url,
			artwork_url = EXCLUDED.artwork_url,
			track_count = EXCLUDED.track_count,
			updated_at = now()
		RETURNING id
	`, pod.SourceID, pod.Title, pod.Author, cats, pod.FeedURL, pod.ArtworkURL, pod.TrackCount).Scan(&id)
	return id, err
}

// SetPinned updates the flag; ErrNotFound if id doesn't exist.
func (p *Postgres) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE podcasts SET pinned = $2, updated_at = now() WHERE id = $1
	`, id, pinned)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateSyncRun inserts a 'running' row (subject is usually the search query).
func (p *Postgres) CreateSyncRun(ctx context.Context, subject string) (domain.SyncRun, error) {
	var run domain.SyncRun
	err := p.pool.QueryRow(ctx, `
		INSERT INTO sync_runs (subject, status, records_processed)
		VALUES ($1, 'running', 0)
		RETURNING id, subject, status, records_processed, error_message, started_at, completed_at
	`, subject).Scan(
		&run.ID, &run.Subject, &run.Status, &run.RecordsProcessed,
		&run.ErrorMessage, &run.StartedAt, &run.CompletedAt,
	)
	return run, err
}

// CompleteSyncRun stamps status, counts, optional error text, and completed_at.
func (p *Postgres) CompleteSyncRun(ctx context.Context, id uuid.UUID, status string, count int, errMsg *string) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE sync_runs
		SET status = $2, records_processed = $3, error_message = $4, completed_at = now()
		WHERE id = $1
	`, id, status, count, errMsg)
	return err
}

// GetSyncRun loads one sync_runs row by id.
func (p *Postgres) GetSyncRun(ctx context.Context, id uuid.UUID) (domain.SyncRun, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT id, subject, status, records_processed, error_message, started_at, completed_at
		FROM sync_runs WHERE id = $1
	`, id)
	var run domain.SyncRun
	err := row.Scan(
		&run.ID, &run.Subject, &run.Status, &run.RecordsProcessed,
		&run.ErrorMessage, &run.StartedAt, &run.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SyncRun{}, ErrNotFound
	}
	return run, err
}

// InsertAudit appends one audit_logs row (metadata defaults to {} if empty).
func (p *Postgres) InsertAudit(ctx context.Context, action, entityID string, metadata json.RawMessage) error {
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO audit_logs (action, entity_id, metadata) VALUES ($1, $2, $3)
	`, action, entityID, metadata)
	return err
}

// ListAuditLogs returns newest-first; clamps silly limits to something reasonable.
func (p *Postgres) ListAuditLogs(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := p.pool.Query(ctx, `
		SELECT id, action, entity_id, metadata, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.AuditLog, 0)
	for rows.Next() {
		var a domain.AuditLog
		if err := rows.Scan(&a.ID, &a.Action, &a.EntityID, &a.Metadata, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// scanPodcasts drains a result set into a slice (shared by list queries).
func scanPodcasts(rows pgx.Rows) ([]domain.Podcast, error) {
	out := make([]domain.Podcast, 0)
	for rows.Next() {
		p, err := scanPodcast(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// scanPodcast maps one row into domain.Podcast (JSON categories + nullable track_count).
func scanPodcast(rows pgx.Row) (domain.Podcast, error) {
	var p domain.Podcast
	var rawCats []byte
	var tc *int32
	err := rows.Scan(
		&p.ID, &p.SourceID, &p.Title, &p.Author, &rawCats,
		&p.FeedURL, &p.ArtworkURL, &tc,
		&p.Pinned, &p.Featured, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return domain.Podcast{}, err
	}
	if len(rawCats) > 0 {
		_ = json.Unmarshal(rawCats, &p.Categories)
	}
	if tc != nil {
		n := int(*tc)
		p.TrackCount = &n
	}
	return p, nil
}

// scanPodcastRow is a thin alias so QueryRow call sites read clearly.
func scanPodcastRow(row pgx.Row) (domain.Podcast, error) {
	return scanPodcast(row)
}
