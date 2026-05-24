package ai

import (
	"strings"

	"github.com/shreyafeo/content-control-plane/internal/config"
)

// FromConfig returns mock when AI_MOCK is true or when live mode is misconfigured.
func FromConfig(cfg config.Config) Suggester {
	if cfg.AIMock {
		return NewMock()
	}
	if cfg.AIProvider == "openai" && strings.TrimSpace(cfg.OpenAIAPIKey) != "" {
		return NewOpenAI(cfg.OpenAIAPIKey, cfg.AIModel, cfg.AITimeout)
	}
	return NewMock()
}
