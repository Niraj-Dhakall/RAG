package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	
)


type SearchResult struct {
	ID string
	Title string
	Snippet string
	Score float64
}


func similaritySearch(ctx context.Context, pool *pgxpool.Pool, embedding []float32, topK int)([]SearchResult, error){
	// <=> means cosine distance
	query := `SELECT id, title, text, 1 - (embedding <=> $1) as score 
			  FROM documents
			  ORDER BY embedding <=> $1
			  LIMIT $2`

	rows, err := pool.Query(ctx, query,pgvector.NewVector(embedding), topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []SearchResult

	for rows.Next(){
		var r SearchResult
		err := rows.Scan(
			&r.ID,
			&r.Title,
			&r.Snippet,
			&r.Score,
		)
		if err != nil{
			return nil, err
		}
		res = append(res, r)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return res , nil

	
}

func connectDB(ctx context.Context)(*pgxpool.Pool, error){
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