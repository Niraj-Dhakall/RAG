package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"rag-project/internal/embeddings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)


type Article struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Text string `json:"text"`
}

// handle the ingest of the articles
// read from articls.jsonl, embed and upsert to db
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
		_, err = pool.Exec(ctx, query, article.ID, article.Title, article.Text, pgVec)
		if err != nil {
			return err
		}
		fmt.Printf("ingested: %s\n", article.ID)

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

	if err := ingestHandler(ctx, pool, "data/articles.jsonl"); err != nil {
		panic(err)
	}
}