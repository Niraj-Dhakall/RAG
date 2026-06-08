package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

/*
This file serves as a way to get a pool connection to the DB, and also perform a similaritySearch with a cosine distance score.
*/

type SearchResult struct {
	ID      string
	Title   string
	Snippet string
	Score   float64
}

// Returns an array of SearchResults and any errors for keyword search
func keywordSearch(ctx context.Context, pool *pgxpool.Pool, query string, topK int) ([]SearchResult, error) {
	sqlQuery := `SELECT id, title, text, ts_rank(ts, websearch_to_tsquery('english', $1)) AS score
			  FROM documents
			  ORDER BY score DESC
			  LIMIT $2`
	rows, err := pool.Query(ctx, sqlQuery, query, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []SearchResult
	for rows.Next() {
		var r SearchResult
		err := rows.Scan(
			&r.ID,
			&r.Title,
			&r.Snippet,
			&r.Score,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	fmt.Printf("Length of result from keyword search: %d", len(res))
	return res, rows.Err()
}

// Returns an array of SearchResults and any errors
func similaritySearch(ctx context.Context, pool *pgxpool.Pool, embedding []float32, topK int) ([]SearchResult, error) {
	// <=> means cosine distance
	query := `SELECT id, title, text, 1 - (embedding <=> $1) as score 
			  FROM documents
			  ORDER BY embedding <=> $1
			  LIMIT $2`

	rows, err := pool.Query(ctx, query, pgvector.NewVector(embedding), topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []SearchResult

	for rows.Next() {
		var r SearchResult
		err := rows.Scan(
			&r.ID,
			&r.Title,
			&r.Snippet,
			&r.Score,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return res, nil

}

// Returns a pointer to the pool connection and any errors
func connectDB(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil

}
