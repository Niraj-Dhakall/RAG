package main

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"rag-project/internal/embeddings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// What this file does:
// Reads articles.jsonl, generates an embedding for each article,
// and upserts the result into the pgvector documents table.

// Step 1: Define an Article struct that matches the JSONL shape:
//   { "id": "wiki_001", "title": "General relativity", "text": "..." }

// Step 2: Connect to Postgres using POSTGRES_DSN (same as api/db.go).
//   Use pgxpool or a plain pgx connection — either works for a one-shot script.
//   Register pgvector types on the connection so you can insert vector values.
//   (Use github.com/pgvector/pgvector-go and pgvector.NewVector())

// Step 3: Open ingest/data/articles.jsonl and decode it line by line.
//   Use a json.Decoder in a loop — each line is one JSON object.

// Step 4: For each article, call the embedding function to get a []float32.
//   You can copy embeddingGenerator() from api/embeddings.go or extract it
//   into a shared package. Easiest: just duplicate the logic here for now.

// Step 5: Upsert the article into the documents table.
//   SQL:
//     INSERT INTO documents (id, title, text, embedding)
//     VALUES ($1, $2, $3, $4)
//     ON CONFLICT (id) DO UPDATE
//       SET title = EXCLUDED.title,
//           text  = EXCLUDED.text,
//           embedding = EXCLUDED.embedding;
//   Pass pgvector.NewVector(embedding) as the $4 argument.

// Step 6: Print progress so you can see it working, e.g.:
//   fmt.Printf("ingested %d/%d: %s\n", i+1, total, article.ID)

// Step 7: main() wires it all together:
//   - load env (os.Getenv or a .env loader)
//   - connect to DB
//   - open the JSONL file
//   - loop through articles calling embed + upsert
//   - print final count or any errors

type Article struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Text string `json:"text"`
}

func ingestHandler(ctx context.Context, pool *pgxpool.Pool, filepath string) error{
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var article Article
		line := scanner.Bytes()

		err := json.Unmarshal(line, &article)
		if err != nil {
			return err
		}
		input := article.Title + "\n" + article.Text
		embedding, err := embeddings.Generate(ctx, input)
		if err != nil {
			return err
		}
		
		pgVec := pgvector.NewVector(embedding)
		query := `INSERT INTO documents (id, title, text, embedding)
				  VALUES ($1, $2, $3, $4)
				  ON CONFLICT (id) DO UPDATE
				  SET title = EXCLUDED.title,
				  	  text = EXCLUDED.text,
					  embedding = EXCLUDED.embedding `
		_ , err = pool.Exec(ctx, query, article.ID, article.Title, article.Text, pgVec)
		if err != nil {
			return  err
		}

	}

	return scanner.Err()

}

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("POSTGRES_DSN"))
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := ingestHandler(ctx, pool, "ingest/data/articles.jsonl"); err != nil {
		panic(err)
	}
}