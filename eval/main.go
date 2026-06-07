package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Query struct {
	Query         string   `json:"query"`
	ExpectedID    string   `json:"expected_id"`
	HardNegatives []string `json:"hard_negatives"`
}

type SearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}

type SearchResult struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}


// handles the evaluation of the RAG system
// Reads queries.jsonl and computes recall@1, recall@3, recall@5, and confusion rate
func evalHandler(ctx context.Context, filepath string) error {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://api:8080"
	}
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var recall1 int
	var recall3 int
	var recall5 int
	var total int
	var confusedTotal int
	// read all queries
	for scanner.Scan() {

		var query Query
		line := scanner.Bytes()
		err := json.Unmarshal(line, &query)
		if err != nil {
			return err
		}
		reqBody := SearchRequest{
			Query: query.Query,
			TopK:  5,
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		// make the search call for each query
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/search", bytes.NewBuffer(body))

		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		var searchResp SearchResponse

		err = json.NewDecoder(res.Body).Decode(&searchResp)
		if err != nil {
			return err
		}
		// evaluate the result for each query
		confused := false
		found := -1
		for i, result := range searchResp.Results {
			if result.ID == query.ExpectedID {
				found = i
				break
			}

			for _, hn := range query.HardNegatives {
				if result.ID == hn {
					confused = true
					break
				}
			}
			if confused {
				confusedTotal++
			}
		}
		if found == 0 {
			recall1++
		}
		if found >= 0 && found < 3 {
			recall3++
		}
		if found >= 0 && found < 5 {
			recall5++
		}
		total++

		res.Body.Close()
	}
	fmt.Printf("Recall@1: %.2f%%\n", float64(recall1)/float64(total)*100)
	fmt.Printf("Recall@3: %.2f%%\n", float64(recall3)/float64(total)*100)
	fmt.Printf("Recall@5: %.2f%%\n", float64(recall5)/float64(total)*100)
	fmt.Printf("Hard-negative confusion rate: %.2f%%\n", float64(confusedTotal)/float64(total)*100)
	fmt.Printf("Total queries processed: %v\n", total)

	return scanner.Err()
}

func main() {
	ctx := context.Background()
	err := evalHandler(ctx, "queries.jsonl")
	if err != nil {
		panic(err)
	}
}
