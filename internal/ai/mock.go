package ai

import (
	"context"
	"strings"

	"github.com/shreyafeo/content-control-plane/internal/domain"
)

type mockSuggester struct{}

// NewMock returns a deterministic suggester for offline demos and CI.
func NewMock() Suggester { return mockSuggester{} }

func (mockSuggester) SuggestMetadata(ctx context.Context, in Input) (Result, error) {
	_ = ctx
	tags := []string{"demo", "ai-suggestion"}
	titleL := strings.ToLower(strings.TrimSpace(in.Title))
	if strings.Contains(titleL, "news") {
		tags = append(tags, "news")
	}
	if strings.Contains(titleL, "tech") {
		tags = append(tags, "technology")
	}
	sum := "Suggested summary (mock): " + in.Title
	if len(sum) > 280 {
		sum = sum[:280]
	}
	return Result{
		Suggestion: domain.MetadataSuggestion{
			Summary:      sum,
			OperatorTags: tags,
			Language:     "en",
			Confidence:   0.42,
		},
		LatencyMS: 2,
		Model:     "mock",
		Provider:  "mock",
	}, nil
}
