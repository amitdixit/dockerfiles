package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type SanctionsList struct {
	XMLName     xml.Name     `xml:"CONSOLIDATED_LIST"`
	Individuals []Individual `xml:"INDIVIDUALS>INDIVIDUAL"`
	Entities    []Entity     `xml:"ENTITIES>ENTITY"`
}

// Individual Sanctioned Person
type Individual struct {
	Type          string
	DataId        string      `xml:"DATAID"`
	FirstName     string      `xml:"FIRST_NAME"`
	SecondName    string      `xml:"SECOND_NAME"`
	ThirdName     string      `xml:"THIRD_NAME"`
	Aliases       []Alias     `xml:"INDIVIDUAL_ALIAS"`
	Documents     []Documents `xml:"INDIVIDUAL_DOCUMENT"`
	Nationalities Value       `xml:"NATIONALITY"`
	DateOfBirth   []Dob       `xml:"INDIVIDUAL_DATE_OF_BIRTH"`
	Country       []Country   `xml:"INDIVIDUAL_ADDRESS"`
	RefercenceNo  string      `xml:"REFERENCE_NUMBER"`
}

// Entity (Organization)
type Entity struct {
	Type         string
	DataId       string    `xml:"DATAID"`
	Name         string    `xml:"FIRST_NAME"`
	Aliases      []Alias   `xml:"ENTITY_ALIAS"`
	Country      []Country `xml:"ENTITY_ADDRESS"`
	RefercenceNo string    `xml:"REFERENCE_NUMBER"`
}

type Documents struct {
	DocumentType   string `xml:"TYPE_OF_DOCUMENT"`
	DocumentNumber string `xml:"NUMBER"`
}

type Alias struct {
	AliasName   string `xml:"ALIAS_NAME"`
	DateOfBirth string `xml:"DATE_OF_BIRTH"`
}

type Country struct {
	CountryName string `xml:"COUNTRY"`
}

type Dob struct {
	Date     string `xml:"DATE"`
	Year     string `xml:"YEAR"`
	FromYear string `xml:"FROM_YEAR"`
	ToYear   string `xml:"TO_YEAR"`
}

// Common value structure
type Value struct {
	Value []string `xml:"VALUE"`
}

func upload(es *elasticsearch.Client) {
	start := time.Now()

	// Open the XML file
	xmlFile, err := os.Open("consolidated_unsc.xml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	// Decode XML
	var list SanctionsList
	decoder := xml.NewDecoder(xmlFile)
	err = decoder.Decode(&list)
	if err != nil {
		fmt.Println("Error decoding XML:", err)
		return
	}

	individualCount := len(list.Individuals)
	entityCount := len(list.Entities)

	fmt.Printf("Individuals: %d\n", individualCount)
	fmt.Printf("Entities: %d\n", entityCount)


	// Index Individuals
	for _, individual := range list.Individuals {
		// fullName := individual.FirstName + " " + individual.SecondName + " " + individual.ThirdName

		var aliases []string
		var dob []string
		for _, alias := range individual.Aliases {
			aliases = append(aliases, strings.TrimSpace(alias.AliasName))
			if alias.DateOfBirth != "" {
				dob = append(dob, alias.DateOfBirth)
			}
		}
		for _, db := range individual.DateOfBirth {
			if db.Date != "" {
				dob = append(dob, db.Date)
			}
			if db.Year != "" {
				dob = append(dob, db.Year)
			}

			if db.FromYear != "" && db.ToYear != "" {
				fromYear, _ := strconv.Atoi(db.FromYear)
				toYear, _ := strconv.Atoi(db.ToYear)
				for i := fromYear; i <= toYear; i++ {
					dob = append(dob, fmt.Sprintf("%d", i))
				} // dob = append(dob, db.FromYear)
			}
		}

		var nationalities []string
		for _, nationality := range individual.Nationalities.Value {
			nationalities = append(nationalities, nationality)
		}

		var countries []string
		countryMap := make(map[string]bool)
		for _, country := range individual.Country {
			if country.CountryName == "" {
				continue
			}
			if !countryMap[country.CountryName] {
				countries = append(countries, country.CountryName)
				countryMap[country.CountryName] = true
			}
		}

		// documentMap := make(map[string]string)
		// for _, country := range individual.Documents {
		// 	if country.CountryName == "" {
		// 		continue
		// 	}
		// 	if !countryMap[country.CountryName] {
		// 		countries = append(countries, country.CountryName)
		// 		countryMap[country.CountryName] = true
		// 	}
		// }

		doc := map[string]interface{}{
			"data_id":       individual.DataId,
			"type":          "individual",
			"name":          fmt.Sprintf("%s %s %s", individual.FirstName, individual.SecondName, individual.ThirdName),
			"aliases":       aliases,
			"date_of_birth": dob,
			"nationalities": nationalities,
			"countries":     countries,
			"refercence_no": individual.RefercenceNo,
			"documents":     individual.Documents,
		}

		fmt.Printf("%+v\n", doc)

		indexData(es, "sanctions_unsc", doc)
	}

	// Index Entities
	for _, entity := range list.Entities {
		var aliases []string
		for _, alias := range entity.Aliases {
			aliases = append(aliases, alias.AliasName)
		}

		var countries []string
		countryMap := make(map[string]bool)
		for _, country := range entity.Country {
			if country.CountryName == "" {
				continue
			}
			if !countryMap[country.CountryName] {
				countries = append(countries, country.CountryName)
				countryMap[country.CountryName] = true
			}
		}

		doc := map[string]interface{}{
			"data_id":       entity.DataId,
			"name":          entity.Name,
			"type":          "entity",
			"aliases":       aliases,
			"countries":     countries,
			"refercence_no": entity.RefercenceNo,
		}

		indexData(es, "sanctions_unsc", doc)
		fmt.Printf("%+v\n", doc)

	}

	elapsed := time.Since(start)
	log.Printf("Data Uploaded successfully in %s", elapsed)
}

func indexData(es *elasticsearch.Client, indexName string, doc map[string]interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		log.Fatalf("Error encoding document: %s", err)
	}

	// res, err := es.Index(
	// 	"sanctions_unsc",
	// 	&buf,
	// 	es.Index.WithContext(context.Background()),
	// )

	res, err := es.Index(
		indexName,
		&buf,
		es.Index.WithContext(context.Background()),
		es.Index.WithRefresh("true"),
	)
	if err != nil {
		log.Fatalf("Error indexing document: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document", res.Status())
	} else {
		log.Printf("[%s] Document indexed successfully.", res.Status())
	}
}
