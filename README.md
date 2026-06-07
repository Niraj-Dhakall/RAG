# RAG Project

A local Retrieval-Augmented Generation (RAG) system. Uses real embeddings, pgvector for vector storage, and a Go HTTP endpoint, all wired together with docker-compose.

## Stack

| Layer | Choice |
|---|---|
| Embeddings | OpenAI `text-embedding-3-small` (or Gemini `embedding-001`) |
| Vector store | PostgreSQL + pgvector |
| Backend | Go (`net/http` + `pgx`) |
| Dataset | ~75 Wikipedia article abstracts |
| Eval | 46 hand-written query/expected-doc pairs |
| Runtime | docker-compose (local only) |

## Project Layout

```

|- api/
│   - main.go          # HTTP server — /search and /health handlers
│   - embeddings.go    # Calls internal embeddings.go to generate embeddings
│   - db.go            # pgvector queries via pgx
│
│- db/
│   - migrations/
│      |- 001_init.sql  # Initializes the documents table
|
|- eval/
│   - main.go          # Runs 46 query/expected-doc pairs (with hard negatives), reports recall@k
│   - queries.jsonl    # Eval set
|
| - ingest/
│    - main.go          # Loads documents, embeds, inserts into DB
│    - data/
│        |- articles.jsonl   # ~75 Wikipedia abstracts
|
|- internal/
│    - embeddings/         
│       |- embeddings.go # Util file to generate embeddings
|
|- screenshots/          # Folder for project screenshots
|- .env.example
|- docker-compose.yml
|-  go.mod
|-  go.sum
|-  README.md
```

## Quickstart

**1. Clone and configure**

```bash
git clone <this-repo>
cd rag-project
cp .env.example .env
# Add OPENAI_API_KEY (or GEMINI_API_KEY) to .env
```

**2. Start services**

```bash
docker-compose up --build
```

This starts:
- `db` - Postgres 15 with the pgvector extension
- `api` - Go HTTP server on `http://localhost:8080`

**3. Ingest documents**

*Make sure `articles.jsonl` has data before building. Will automatically ingest once built*

Reads `ingest/data/articles.jsonl`, generates embeddings, and upserts into pgvector. Takes ~30 seconds for 75 documents.

**4. Query**

```bash
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{"query": "black holes and general relativity", "top_k": 5}'
```

**5. Run eval**

```bash
cd /eval && go run main.go
```

Reports `recall@1`, `recall@3`, and `recall@5` across 46 hand-written query/expected-doc pairs.

## API

### `POST /search`

```json
{
  "query": "string",
  "top_k": 5
}
```

Returns the top-k most similar documents by cosine similarity.

```json
{
  "results": [
    {
      "id": "wiki_042",
      "title": "Black hole",
      "snippet": "A black hole is a region of spacetime...",
      "score": 0.91
    }
  ]
}
```

### `GET /health`

Returns `{"status": "ok"}` when the API and DB are reachable.

## Dataset

~75 abstracts pulled from the Wikipedia featured articles list, covering physics, biology, history, and mathematics. Stored as newline-delimited JSON in `ingest/data/articles.jsonl`:

```json
{"id": "wiki_001", "title": "General relativity", "text": "General relativity is the geometric theory of gravitation..."}
```

## Eval

46 hand-written pairs in `eval/queries.json`:

```json
[
  {
    "query": "Einstein's theory of gravity",
    "expected_id": "wiki_001",
    "hard_negatives": ["wiki_###"...],
    "style": "student"
  }
]
```

Sample results:

| Metric | Score |
|---|---|
| recall@1 | 93.48% |
| recall@3 | 100.00% |
| recall@5 | 100.00% |
|Hard-negative confusion rate|  6.52% | 
|Total queries processed| 46|
## Environment Variables

```
OPENAI_API_KEY=...        # Required if using OpenAI embeddings
GEMINI_API_KEY=...           # Required if using Gemini embeddings
EMBEDDING_PROVIDER=    # "openai" or "gemini"
POSTGRES_DSN=postgresql://rag:rag@postgres:5432/rag
POSTGRES_USER=rag
POSTGRES_PASSWORD=rag
POSTGRES_DB=rag
API_URL=http://localhost:8080
```

## Screenshots
![Search result2](screenshots/APICall2S.png)
![Search results](screenshots/APICallS.png)
![Database](screenshots/DatabaseS.png)

## Notes

- No cloud deployment, runs entirely on `docker-compose` locally.
- pgvector uses an IVFFlat index (`lists=50`) for fast ANN search at this dataset size.
- Swap `EMBEDDING_PROVIDER` in `.env` to switch between OpenAI and Gemini with no code changes.
