CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE skills (
    id          SERIAL PRIMARY KEY,
    slug        VARCHAR(64)  UNIQUE NOT NULL,
    name        VARCHAR(128) NOT NULL,
    category    INTEGER      NOT NULL DEFAULT 0,
    popularity  INTEGER      NOT NULL DEFAULT 0,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at  TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_skills_slug_trgm  ON skills USING GIN (slug gin_trgm_ops);
CREATE INDEX idx_skills_name_trgm  ON skills USING GIN (name gin_trgm_ops);
CREATE INDEX idx_skills_category   ON skills(category);
CREATE INDEX idx_skills_popularity ON skills(popularity DESC);
