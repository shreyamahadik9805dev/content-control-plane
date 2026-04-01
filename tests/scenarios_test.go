package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/shreyafeo/content-control-plane/internal/cache"
	"github.com/shreyafeo/content-control-plane/internal/client/itunes"
	"github.com/shreyafeo/content-control-plane/internal/handler"
	"github.com/shreyafeo/content-control-plane/internal/repository"
	"github.com/shreyafeo/content-control-plane/internal/service"
)

// Flow tests: run with go test ./tests/... -v

// Mock iTunes still returns plausible rows so local / Docker runs do not need Apple’s network.
func TestFlow_MockITunesWorksOffline(t *testing.T) {
	c := itunes.New("", true)
	got, err := c.SearchPodcasts(context.Background(), "anything")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 1 {
		t.Fatal("expected at least one mock show")
	}
	if got[0].CollectionID == 0 || got[0].CollectionName == "" {
		t.Fatalf("mock row looks broken: %+v", got[0])
	}
}

// Whitespace-only query is treated as “missing”; we bail with ErrBadQuery before touching the repo.
func TestFlow_SyncRejectsEmptyQuery(t *testing.T) {
	s := service.NewPodcasts(&noopStore{}, cache.New(time.Minute, 2*time.Minute), itunes.New("", true))
	_, err := s.Sync(context.Background(), "   ")
	if err != service.ErrBadQuery {
		t.Fatalf("want ErrBadQuery, got %v", err)
	}
}

// Full sync with mock iTunes + fake DB: we expect a successful run and one upsert per mock show.
func TestFlow_SyncUpsertsMockShows(t *testing.T) {
	st := &stubStore{}
	s := service.NewPodcasts(st, cache.New(time.Minute, 2*time.Minute), itunes.New("", true))

	run, err := s.Sync(context.Background(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != "success" {
		t.Fatalf("status: %q", run.Status)
	}
	if st.upsertCalls != 2 {
		t.Fatalf("upserts: want 2 (mock catalog size), got %d", st.upsertCalls)
	}
}

// Liveness stays cheap: /health is 200 even when the podcast service is nil.
func TestFlow_LivenessAlwaysOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.New(nil)
	r := gin.New()
	h.Register(r)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
}

// POST /sync/podcasts with no ?query= is a client mistake (400), not an upstream failure (502).
func TestFlow_SyncWithoutQueryIs400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := service.NewPodcasts(&noopStore{}, cache.New(time.Minute, 2*time.Minute), itunes.New("", true))
	h := handler.New(svc)

	r := gin.New()
	h.Register(r)

	req := httptest.NewRequest(http.MethodPost, "/sync/podcasts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
	}
}

// Repo “not found” errors are recognizable so handlers can map them to 404 without string matching.
func TestFlow_ErrIsNotFoundForHTTPLayer(t *testing.T) {
	if !service.ErrIsNotFound(repository.ErrNotFound) {
		t.Fatal("expected true for repo ErrNotFound")
	}
	if service.ErrIsNotFound(context.Canceled) {
		t.Fatal("expected false for unrelated error")
	}
}
