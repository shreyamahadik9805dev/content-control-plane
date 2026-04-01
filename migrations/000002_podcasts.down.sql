DROP TABLE IF EXISTS podcasts;

CREATE TABLE books (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id text NOT NULL UNIQUE,
    title text NOT NULL,
    author text NOT NULL,
    subjects jsonb NOT NULL DEFAULT '[]'::jsonb,
    publish_year int,
    pinned boolean NOT NULL DEFAULT false,
    featured boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_books_updated_at ON books (updated_at DESC);
