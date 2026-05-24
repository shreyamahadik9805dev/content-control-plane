package service

import (
	"testing"

	"github.com/shreyafeo/content-control-plane/internal/domain"
)

func TestNormalizeSuggestion_TrimsAndCaps(t *testing.T) {
	ms := domain.MetadataSuggestion{
		Summary:      "  hello  ",
		OperatorTags: []string{"News", "news", "Tech"},
		Confidence:   2.0,
		Language:     "en-US-extra",
	}
	normalizeSuggestion(&ms)
	if ms.Summary != "hello" {
		t.Fatalf("summary: %q", ms.Summary)
	}
	if ms.Confidence != 1 {
		t.Fatalf("confidence capped: %v", ms.Confidence)
	}
	if len(ms.OperatorTags) != 2 || ms.OperatorTags[0] != "news" {
		t.Fatalf("dedupe tags: %v", ms.OperatorTags)
	}
	if ms.Language != "en-US-ex" {
		t.Fatalf("language trim: %q", ms.Language)
	}
}
