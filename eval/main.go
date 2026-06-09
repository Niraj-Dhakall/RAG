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
func evalHandler(ctx context.Context, filepath string, endpoint string) (recall1, recall3, recall5, confused, total int, err error) {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://api:8080"
	}
	file, err := os.Open(filepath)
	if err != nil {
		return 0, 0, 0, 0, 0, err // idk how to get rid of needing to do this to return an error
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	var confusedTotal int
	// read all queries
	for scanner.Scan() {

		var query Query
		line := scanner.Bytes()
		err := json.Unmarshal(line, &query)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}
		reqBody := SearchRequest{
			Query: query.Query,
			TopK:  5,
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}
		// make the search call for each query
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+endpoint, bytes.NewBuffer(body))

		if err != nil {
			return 0, 0, 0, 0, 0, err
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}

		var searchResp SearchResponse

		err = json.NewDecoder(res.Body).Decode(&searchResp)
		if err != nil {
			return 0, 0, 0, 0, 0, err
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

	err = scanner.Err()
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	return recall1, recall3, recall5, confusedTotal, total, nil
}

func main() {
	ctx := context.Background()

	semR1, semR3, semR5, semC, semT, err := evalHandler(ctx, "queries.jsonl", "/search")
	if err != nil {
		panic(err)
	}
	keyR1, keyR3, keyR5, keyC, keyT, err := evalHandler(ctx, "queries.jsonl", "/search/keyword")
	if err != nil {
		panic(err)
	}
	hybR1, hybR3, hybR5, hybC, hybT, err := evalHandler(ctx, "queries.jsonl", "/search/hybrid")
	if err != nil {
		panic(err)
	}

	lines := []string{
		fmt.Sprintf("%-18s %-12s %-12s %-12s", "", "Semantic", "Keyword", "Hybrid"),
		fmt.Sprintf("%-18s %-12.2f%% %-12.2f%% %-12.2f%%", "Recall@1", float64(semR1)/float64(semT)*100, float64(keyR1)/float64(keyT)*100, float64(hybR1)/float64(hybT)*100),
		fmt.Sprintf("%-18s %-12.2f%% %-12.2f%% %-12.2f%%", "Recall@3", float64(semR3)/float64(semT)*100, float64(keyR3)/float64(keyT)*100, float64(hybR3)/float64(hybT)*100),
		fmt.Sprintf("%-18s %-12.2f%% %-12.2f%% %-12.2f%%", "Recall@5", float64(semR5)/float64(semT)*100, float64(keyR5)/float64(keyT)*100, float64(hybR5)/float64(hybT)*100),
		fmt.Sprintf("%-18s %-12.2f%% %-12.2f%% %-12.2f%%", "Confusion rate", float64(semC)/float64(semT)*100, float64(keyC)/float64(keyT)*100, float64(hybC)/float64(hybT)*100),
	}
	for _, l := range lines {
		fmt.Println(l)
	}
	err = os.MkdirAll("/results", 0755)
	if err != nil {
		panic(err)
	}
	err = os.Remove("/results/eval_results.txt")
	if err != nil {
		panic(err)
	}
	out, err := os.Create("/results/eval_results.txt")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	for _, l := range lines {
		fmt.Fprintln(out, l)
	}

}
