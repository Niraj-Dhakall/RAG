CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE TABLE documents (
	id TEXT PRIMARY KEY,
	title TEXT,
	text TEXT,
	embedding VECTOR(1536),
	ts tsvector GENERATED ALWAYS AS (
		to_tsvector('english'::regconfig, title || ' ' || text)
	) STORED
);

CREATE INDEX ON documents USING hnsw (embedding vector_cosine_ops);
CREATE INDEX ON documents USING GIN(ts);
CREATE INDEX ON documents USING GIN((title || ' '|| text) gin_trgm_ops)