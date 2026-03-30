package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/shreyafeo/content-control-plane/internal/cache"
	"github.com/shreyafeo/content-control-plane/internal/client/itunes"
	"github.com/shreyafeo/content-control-plane/internal/domain"
	"github.com/shreyafeo/content-control-plane/internal/repository"
)

const cacheKeyList = "podcasts:list"

var ErrBadQuery = errors.New("query is required")

type Podcasts struct {
	repo   repository.Store
	cache  *cache.TTL
	itunes *itunes.Client
}

func NewPodcasts(repo repository.Store, c *cache.TTL, client *itunes.Client) *Podcasts {
	return &Podcasts{repo: repo, cache: c, itunes: client}
}

func (s *Podcasts) List(ctx context.Context) ([]domain.Podcast, error) {
	if v, ok := s.cache.Get(cacheKeyList); ok {
		if pods, ok := v.([]domain.Podcast); ok {
			return pods, nil
		}
	}
	pods, err := s.repo.ListPodcasts(ctx)
	if err != nil {
		return nil, err
	}
	s.cache.Set(cacheKeyList, pods)
	return pods, nil
}

func (s *Podcasts) Get(ctx context.Context, id uuid.UUID) (domain.Podcast, error) {
	key := "podcast:" + id.String()
	if v, ok := s.cache.Get(key); ok {
		if p, ok := v.(domain.Podcast); ok {
			return p, nil
		}
	}
	p, err := s.repo.GetPodcastByID(ctx, id)
	if err != nil {
		return domain.Podcast{}, err
	}
	s.cache.Set(key, p)
	return p, nil
}

type PinRequest struct {
	Pinned bool `json:"pinned"`
}

func (s *Podcasts) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error {
	if err := s.repo.SetPinned(ctx, id, pinned); err != nil {
		return err
	}
	s.cache.Delete("podcast:" + id.String())
	s.cache.Delete(cacheKeyList)
	meta, _ := json.Marshal(map[string]bool{"pinned": pinned})
	action := "podcast.unpinned"
	if pinned {
		action = "podcast.pinned"
	}
	return s.repo.InsertAudit(ctx, action, id.String(), meta)
}

func (s *Podcasts) Sync(ctx context.Context, query string) (domain.SyncRun, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return domain.SyncRun{}, ErrBadQuery
	}

	run, err := s.repo.CreateSyncRun(ctx, query)
	if err != nil {
		return domain.SyncRun{}, err
	}

	shows, err := s.itunes.SearchPodcasts(ctx, query)
	if err != nil {
		msg := err.Error()
		_ = s.repo.CompleteSyncRun(ctx, run.ID, "failed", 0, &msg)
		em, _ := json.Marshal(map[string]string{"error": msg})
		_ = s.repo.InsertAudit(ctx, "sync.failed", run.ID.String(), em)
		return domain.SyncRun{}, err
	}

	count := 0
	for _, sh := range shows {
		p := normalizeShow(sh)
		if p.SourceID == "" || p.Title == "" {
			continue
		}
		if _, err := s.repo.UpsertPodcast(ctx, p); err != nil {
			msg := err.Error()
			_ = s.repo.CompleteSyncRun(ctx, run.ID, "failed", count, &msg)
			return domain.SyncRun{}, err
		}
		count++
	}

	if err := s.repo.CompleteSyncRun(ctx, run.ID, "success", count, nil); err != nil {
		return domain.SyncRun{}, err
	}

	meta, _ := json.Marshal(map[string]any{
		"query":   query,
		"records": count,
	})
	_ = s.repo.InsertAudit(ctx, "sync.completed", run.ID.String(), meta)
	s.cache.Delete(cacheKeyList)

	out, err := s.repo.GetSyncRun(ctx, run.ID)
	if err != nil {
		return domain.SyncRun{}, err
	}
	return out, nil
}

func (s *Podcasts) AuditLogs(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	return s.repo.ListAuditLogs(ctx, limit)
}

func normalizeShow(sh itunes.Show) domain.Podcast {
	if sh.CollectionID == 0 {
		return domain.Podcast{}
	}
	sourceID := strconv.FormatInt(sh.CollectionID, 10)
	title := strings.TrimSpace(sh.CollectionName)
	author := strings.TrimSpace(sh.ArtistName)
	if author == "" {
		author = "Unknown"
	}
	cats := append([]string(nil), sh.Genres...)
	tc := sh.TrackCount
	var tcp *int
	if tc > 0 {
		tcp = &tc
	}
	return domain.Podcast{
		SourceID:   sourceID,
		Title:      title,
		Author:     author,
		Categories: cats,
		FeedURL:    strings.TrimSpace(sh.FeedURL),
		ArtworkURL: strings.TrimSpace(sh.ArtworkURL600),
		TrackCount: tcp,
	}
}

func ErrIsNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}
