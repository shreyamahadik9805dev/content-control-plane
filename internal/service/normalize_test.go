package service

import (
	"testing"

	"github.com/shreyafeo/content-control-plane/internal/client/itunes"
)

// Mapping tests (Apple JSON → our row). Run: go test ./internal/service -run TestMap_ -v

// We copy the right fields into our shape: stable source id, trimmed title, default author when missing, track count.
func TestMap_AppleJSON_ToCatalogRow(t *testing.T) {
	got := normalizeShow(itunes.Show{
		CollectionID:   42,
		ArtistName:     "",
		CollectionName: "  My Show  ",
		FeedURL:        "https://example.com/feed",
		ArtworkURL600:  "https://example.com/art.png",
		Genres:         []string{"News"},
		TrackCount:     10,
	})
	if got.SourceID != "42" {
		t.Fatalf("source_id: got %q", got.SourceID)
	}
	if got.Title != "My Show" {
		t.Fatalf("title trim: got %q", got.Title)
	}
	if got.Author != "Unknown" {
		t.Fatalf("empty artist should default to Unknown, got %q", got.Author)
	}
	if got.TrackCount == nil || *got.TrackCount != 10 {
		t.Fatalf("track count pointer: %+v", got.TrackCount)
	}
}

// No collection id means we do not fabricate a row—avoids junk in the catalog.
func TestMap_AppleJSON_ZeroCollectionDropped(t *testing.T) {
	got := normalizeShow(itunes.Show{CollectionID: 0, CollectionName: "orphan"})
	if got.SourceID != "" || got.Title != "" {
		t.Fatalf("expected empty podcast, got %+v", got)
	}
}
