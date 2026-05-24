DROP TABLE IF EXISTS ai_proposals;

ALTER TABLE podcasts
    DROP COLUMN IF EXISTS summary,
    DROP COLUMN IF EXISTS operator_tags;
