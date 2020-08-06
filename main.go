package main

import (
	"fmt"
	// "io/ioutil"
	// "html"
	"encoding/base64"
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Iwark/spreadsheet"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func getReputations(token, groupID string) ([][]string, error) {
	client := &http.Client{}
	endpoint := fmt.Sprintf("https://api.campuswire.com/v1/group/%s/reputation_report", groupID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	csvreader := csv.NewReader(resp.Body)

	return csvreader.ReadAll()
}

func sheetClient(secret []byte) (s *spreadsheet.Service, err error) {
	conf, err := google.JWTConfigFromJSON(secret, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return
	}

	s = spreadsheet.NewServiceWithClient(conf.Client(oauth2.NoContext))
	return
}

func main() {
	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	sheetID := os.Getenv("SHEETID")
	clientSecretb64 := os.Getenv("SHEETSECRET")
	clientSecret, _ := base64.StdEncoding.DecodeString(clientSecretb64)
	cwToken := os.Getenv("CWTOKEN")

	cilent, err := sheetClient(clientSecret)
	if err != nil {
		log.Fatalln(err)
	}
	spreadsheet, err := cilent.FetchSpreadsheet(sheetID)
	if err != nil {
		log.Fatalln(err)
	}

	firstsheet, err := spreadsheet.SheetByTitle("course")

	for i, row := range firstsheet.Rows {
		if i == 0 {
			continue
		}
		active := row[0].Value != "0"
		if !active {
			continue
		}
		course := row[1].Value
		groupId := row[2].Value

		coursesheet, err := spreadsheet.SheetByTitle(course)
		if err != nil {
			log.Fatalln(err)
		}

		records, err := getReputations(cwToken, groupId)
		if err != nil {
			log.Fatalln(err)
		}

		n := len(coursesheet.Rows)
		m := len(records[0])

		for i, record := range records {
			if i == 0 {
				continue
			}
			for j, field := range record {
				coursesheet.Update(n+i-1, j, field)
			}
			coursesheet.Update(n+i-1, m, time.Now().Format(time.RFC3339))
		}
		err = coursesheet.Synchronize()
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("course %s updated with %d records\n", course, len(records))
	}
}
