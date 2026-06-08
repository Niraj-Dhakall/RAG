CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE documents (
	id TEXT PRIMARY KEY,
	title TEXT,
	text TEXT,
	embedding VECTOR(1536),
	ts tsvector GENERATED ALWAYS AS (
		to_tsvector('english'::regconfig, title || ' ' || text)
	) STORED
);

CREATE INDEX ON documents USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);
CREATE INDEX ON documents USING GIN(ts);
