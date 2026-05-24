package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/shreyafeo/content-control-plane/internal/ai"
	"github.com/shreyafeo/content-control-plane/internal/domain"
	"github.com/shreyafeo/content-control-plane/internal/repository"
)

var (
	ErrProposalNotPending = errors.New("proposal is not pending")
)

var slugSanitize = regexp.MustCompile(`[^a-z0-9-]+`)

// Proposals orchestrates human-in-the-loop AI suggestions (persistence + apply).
type Proposals struct {
	repo repository.Store
	sg   ai.Suggester
	pods *Podcasts
}

func NewProposals(repo repository.Store, sg ai.Suggester, pods *Podcasts) *Proposals {
	return &Proposals{repo: repo, sg: sg, pods: pods}
}

// Generate creates a pending proposal row; does not change catalog rows.
func (p *Proposals) Generate(ctx context.Context, podcastID uuid.UUID) (domain.AIProposal, error) {
	pod, err := p.repo.GetPodcastByID(ctx, podcastID)
	if err != nil {
		return domain.AIProposal{}, err
	}

	in := ai.Input{
		Title:      pod.Title,
		Author:     pod.Author,
		Categories: pod.Categories,
		FeedURL:    pod.FeedURL,
		SourceID:   pod.SourceID,
		ArtworkURL: pod.ArtworkURL,
		TrackCount: pod.TrackCount,
	}
	ctxJSON, err := json.Marshal(in)
	if err != nil {
		return domain.AIProposal{}, err
	}

	res, err := p.sg.SuggestMetadata(ctx, in)
	if err != nil {
		return domain.AIProposal{}, err
	}
	normalizeSuggestion(&res.Suggestion)

	payload, err := json.Marshal(res.Suggestion)
	if err != nil {
		return domain.AIProposal{}, err
	}

	lat := res.LatencyMS
	pr := domain.AIProposal{
		PodcastID: podcastID,
		Status:    "pending",
		Kind:      "metadata_enrich",
		Payload:   payload,
		Context:   ctxJSON,
		Model:     res.Model,
		Provider:  res.Provider,
		LatencyMS: &lat,
	}

	id, err := p.repo.InsertProposal(ctx, pr)
	if err != nil {
		return domain.AIProposal{}, err
	}

	meta, _ := json.Marshal(map[string]string{
		"proposal_id": id.String(),
		"podcast_id":  podcastID.String(),
		"model":       res.Model,
	})
	_ = p.repo.InsertAudit(ctx, "proposal.created", id.String(), meta)

	out, err := p.repo.GetProposalByID(ctx, id)
	if err != nil {
		return domain.AIProposal{}, err
	}
	return out, nil
}

// ListForPodcast returns proposals for one show; status empty = all.
func (p *Proposals) ListForPodcast(ctx context.Context, podcastID uuid.UUID, status string) ([]domain.AIProposal, error) {
	if _, err := p.repo.GetPodcastByID(ctx, podcastID); err != nil {
		return nil, err
	}
	return p.repo.ListProposalsForPodcast(ctx, podcastID, status)
}

// Accept applies payload to the podcast and marks the proposal accepted.
func (p *Proposals) Accept(ctx context.Context, proposalID uuid.UUID) (domain.Podcast, error) {
	pr, err := p.repo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return domain.Podcast{}, err
	}
	if pr.Status != "pending" {
		return domain.Podcast{}, ErrProposalNotPending
	}

	var ms domain.MetadataSuggestion
	if err := json.Unmarshal(pr.Payload, &ms); err != nil {
		return domain.Podcast{}, err
	}
	normalizeSuggestion(&ms)

	if err := p.pods.ApplyEnrichment(ctx, pr.PodcastID, ms.Summary, ms.OperatorTags); err != nil {
		return domain.Podcast{}, err
	}

	now := time.Now()
	if err := p.repo.SetProposalStatus(ctx, proposalID, "accepted", now); err != nil {
		return domain.Podcast{}, err
	}

	meta, _ := json.Marshal(map[string]string{
		"proposal_id": proposalID.String(),
		"podcast_id":  pr.PodcastID.String(),
		"model":       pr.Model,
	})
	_ = p.repo.InsertAudit(ctx, "proposal.accepted", proposalID.String(), meta)

	return p.pods.Get(ctx, pr.PodcastID)
}

// Reject marks a proposal rejected without changing catalog data.
func (p *Proposals) Reject(ctx context.Context, proposalID uuid.UUID, note string) error {
	pr, err := p.repo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return err
	}
	if pr.Status != "pending" {
		return ErrProposalNotPending
	}

	now := time.Now()
	if err := p.repo.SetProposalStatus(ctx, proposalID, "rejected", now); err != nil {
		return err
	}
	meta, _ := json.Marshal(map[string]string{
		"proposal_id": proposalID.String(),
		"podcast_id":  pr.PodcastID.String(),
		"note":        strings.TrimSpace(note),
	})
	return p.repo.InsertAudit(ctx, "proposal.rejected", proposalID.String(), meta)
}

func normalizeSuggestion(ms *domain.MetadataSuggestion) {
	ms.Summary = strings.TrimSpace(ms.Summary)
	if len(ms.Summary) > 500 {
		ms.Summary = ms.Summary[:500]
	}

	seen := make(map[string]struct{})
	out := make([]string, 0, len(ms.OperatorTags))
	for _, t := range ms.OperatorTags {
		t = strings.ToLower(strings.TrimSpace(t))
		t = slugSanitize.ReplaceAllString(t, "-")
		t = strings.Trim(t, "-")
		if t == "" || len(t) > 32 {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
		if len(out) >= 12 {
			break
		}
	}
	ms.OperatorTags = out

	if ms.Confidence < 0 {
		ms.Confidence = 0
	}
	if ms.Confidence > 1 {
		ms.Confidence = 1
	}
	ms.Language = strings.TrimSpace(ms.Language)
	if len(ms.Language) > 8 {
		ms.Language = ms.Language[:8]
	}
}
