DROP TABLE IF EXISTS books;

CREATE TABLE podcasts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id text NOT NULL UNIQUE,
    title text NOT NULL,
    author text NOT NULL,
    categories jsonb NOT NULL DEFAULT '[]'::jsonb,
    feed_url text NOT NULL DEFAULT '',
    artwork_url text NOT NULL DEFAULT '',
    track_count int,
    pinned boolean NOT NULL DEFAULT false,
    featured boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_podcasts_updated_at ON podcasts (updated_at DESC);
