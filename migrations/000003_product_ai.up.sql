ALTER TABLE podcasts
    ADD COLUMN IF NOT EXISTS summary text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS operator_tags jsonb NOT NULL DEFAULT '[]'::jsonb;

CREATE TABLE ai_proposals (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    podcast_id uuid NOT NULL REFERENCES podcasts(id) ON DELETE CASCADE,
    status text NOT NULL CHECK (status IN ('pending', 'accepted', 'rejected')),
    kind text NOT NULL DEFAULT 'metadata_enrich',
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    context jsonb NOT NULL DEFAULT '{}'::jsonb,
    model text NOT NULL DEFAULT '',
    provider text NOT NULL DEFAULT '',
    latency_ms int,
    created_at timestamptz NOT NULL DEFAULT now(),
    resolved_at timestamptz
);

CREATE INDEX idx_ai_proposals_podcast_created ON ai_proposals (podcast_id, created_at DESC);
