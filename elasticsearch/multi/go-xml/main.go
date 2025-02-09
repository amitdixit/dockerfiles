package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

// Define XML structures

// Common alias structure

func main() {
	// // Create Elasticsearch client
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9201",
			"http://localhost:9202",
			"http://localhost:9203",
		},
	})
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}
	// createIndex(es)
	// upload(es)

	//searchCustomers(es, "Mohammad Tahir")
	// Perform search
		//searchCustomersAll(es, "Muhammad Tahir", "Afghanistan", "1961")

		searchCustomersAll2(es, "Muhammad Tahir", "Afghanistan", "1961","" )


}

// Function to index data into Elasticsearch
















// Structs for Elasticsearch Query
type SearchQuery struct {
	Query BoolQuery `json:"query"`
	Size  int       `json:"size"`
}

type BoolQuery struct {
	Bool BoolConditions `json:"bool"`
}

type BoolConditions struct {
	Should []MatchQuery `json:"should"`
	Filter []TermQuery  `json:"filter"`
}

type MatchQuery struct {
	Match map[string]MatchCondition `json:"match"`
}

type MatchCondition struct {
	Query     string `json:"query"`
	Fuzziness string `json:"fuzziness,omitempty"`
}

type TermQuery struct {
	Term map[string]string `json:"term"`
}
func searchElasticsearch(es *elasticsearch.Client, name, alias, dob, nationality string) {
	// Construct the search query
	query := SearchQuery{
		Query: BoolQuery{
			Bool: BoolConditions{
				Should: []MatchQuery{
					{Match: map[string]MatchCondition{"name": {Query: name, Fuzziness: "AUTO"}}},
					{Match: map[string]MatchCondition{"aliases": {Query: alias, Fuzziness: "AUTO"}}},
				},
				Filter: []TermQuery{
					{Term: map[string]string{"date_of_birth": dob}},
					{Term: map[string]string{"nationality": nationality}},
				},
			},
		},
		Size: 10,
	}

	// Convert query to JSON
	jsonQuery, err := json.Marshal(query)
	if err != nil {
		log.Fatalf("Error marshalling query: %v", err)
	}

	// Perform the search request
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("sanctions_unsc"), // Change to your index name
		es.Search.WithBody(bytes.NewReader(jsonQuery)),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Search request failed: %v", err)
	}
	defer res.Body.Close()

	// Print search response
	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("Search Results:\n", string(resultJSON))
}
