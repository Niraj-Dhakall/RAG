CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE documents (
	id TEXT PRIMARY KEY,
	title TEXT,
	text TEXT,
	embedding VECTOR(1536)
);

CREATE INDEX ON documents USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);