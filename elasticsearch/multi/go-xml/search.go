package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

func createIndex(es *elasticsearch.Client) {
	indexMapping := `{
		"settings": {
			"analysis": {
				"analyzer": {
					"name_analyzer": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "asciifolding", "porter_stem"]
					},
					"phonetic_analyzer": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "asciifolding", "double_metaphone"]
					}
				},
				"filter": {
					"double_metaphone": {
						"type": "phonetic",
						"encoder": "double_metaphone",
						"replace": false
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"type": { "type": "keyword" },
				"data_id": { "type": "keyword" },
				"name": {
					"type": "text",
					"analyzer": "name_analyzer",
					"fields": {
						"phonetic": { "type": "text", "analyzer": "phonetic_analyzer" }
					}
				},
				"aliases": {
					"type": "text",
					"analyzer": "name_analyzer",
					"fields": {
						"phonetic": { "type": "text", "analyzer": "phonetic_analyzer" }
					}
				},
				"date_of_birth": { "type": "date", "format": "yyyy-MM-dd||yyyy" },
				"nationalities": { "type": "keyword" },
				"countries": { "type": "keyword" },
				"documents": {
					"type": "nested",
					"properties": {
						"type": { "type": "keyword" },
						"type2": { "type": "keyword" }, 
						"number": { "type": "keyword" },
						"issuing_country": { "type": "keyword" },
						"date_of_issue": { "type": "date", "format": "yyyy-MM-dd" },
						"city_of_issue": { "type": "keyword" },
						"country_of_issue": { "type": "keyword" },
						"note": { "type": "text" }
						}
				}
			}
		}
	}`

	// Create Index
	res, err := es.Indices.Create("sanctions_unsc", es.Indices.Create.WithBody(strings.NewReader(indexMapping)))
	if err != nil {
		log.Fatalf("Error creating index: %s", err)
	}
	defer res.Body.Close()

	log.Println("Index 'sanctions_unsc' created successfully!")
}


func searchCustomers(es *elasticsearch.Client, query string) {
	searchBody := fmt.Sprintf(`{
		"query": {
			"bool": {
				"should": [
					{ "match": { "name": "%s" } },
					{ "match": { "name.phonetic": "%s" } },
					{ "match": { "aliases": "%s" } },
					{ "match": { "aliases.phonetic": "%s" } }
				]
			}
		}
	}`, query, query, query, query)

	res, err := es.Search(
		es.Search.WithIndex("sanctions_unsc"),
		es.Search.WithBody(strings.NewReader(searchBody)),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error searching: %s", err)
	}
	defer res.Body.Close()

	fmt.Println(res)
}


func searchCustomersAll(es *elasticsearch.Client, name string, nationality string, dob string) {
	searchBody := fmt.Sprintf(`{
		"query": {
			"bool": {
				"must": [
					{
						"bool": {
							"should": [
								{ "match": { "name": "%s" } },
								{ "match": { "name.phonetic": "%s" } },
								{ "fuzzy": { "name": { "value": "%s", "fuzziness": "AUTO" } } },
								{ "match": { "aliases": "%s" } },
								{ "match": { "aliases.phonetic": "%s" } }
							]
						}
					}
				],
				"filter": [
					{ "term": { "nationality": "%s" } },
					{
						"bool": {
							"should": [
								{ "range": { "dob": { "gte": "%s", "lte": "%s" } } },
								{ "range": { "dob": { "gte": "%s-01-01", "lte": "%s-12-31" } } }
							]
						}
					}
				]
			}
		}
	}`, name, name, name, name, name, nationality, dob, dob, dob, dob)

	// Convert searchBody string to JSON
	var bodyBuffer bytes.Buffer
	json.Compact(&bodyBuffer, []byte(searchBody))

	// Execute Elasticsearch query
	res, err := es.Search(
		es.Search.WithIndex("sanctions_unsc"),
		es.Search.WithBody(&bodyBuffer),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error searching: %s", err)
	}
	defer res.Body.Close()

	fmt.Println(res)
}


// Struct for Elasticsearch response parsing
type EsResponse struct {
	Hits struct {
		Hits []struct {
			Score  float64                `json:"_score"`
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
// Function to search customers in Elasticsearch
func searchCustomersAll2(es *elasticsearch.Client, name, nationality, dob, documentNumber string) {
	// Construct the Elasticsearch Query
	query := fmt.Sprintf(`{
		"query": {
			"bool": {
				"should": [
					{ 
						"multi_match": {
							"query": "%s",
							"fields": ["name", "name.phonetic", "aliases", "aliases.phonetic"],
							"fuzziness": "AUTO",
							"boost": 3
						}
					},
					{
						"term": { "nationality": "%s" }
					},
					{
						"range": {
							"dob": {
								"gte": "%s-01-01",
								"lte": "%s-12-31",
								"boost": 2
							}
						}
					},
					{
						"nested": {
							"path": "documents",
							"query": {
								"bool": {
									"should": [
										{ "term": { "documents.number": "%s" } }
									]
								}
							}
						}
					}
				],
				"minimum_should_match": 1
			}
		}
	}`, name, nationality, dob, dob, documentNumber)

	// Execute the search request
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("sanctions_unsc"),
		es.Search.WithBody(strings.NewReader(query)),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error executing search query: %s", err)
	}
	defer res.Body.Close()

	// Parse response
	var esRes EsResponse
	if err := json.NewDecoder(res.Body).Decode(&esRes); err != nil {
		log.Fatalf("Error parsing response: %s", err)
	}

	// Print results
	fmt.Println("\n🔎 Search Results:")
	if len(esRes.Hits.Hits) == 0 {
		fmt.Println("No matches found.")
		return
	}

	for i, hit := range esRes.Hits.Hits {
		fmt.Printf("\nResult #%d - Score: %.2f\n", i+1, hit.Score)
		fmt.Printf("Name: %s\n", hit.Source["name"])
		fmt.Printf("Nationality: %s\n", hit.Source["nationalities"])
		if dob, exists := hit.Source["date_of_birth"]; exists {
			fmt.Printf("DOB: %s\n", dob)
		}
		fmt.Println("--------------------------------------")
	}
}
