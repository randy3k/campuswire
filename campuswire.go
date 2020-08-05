package campuswire

import (
	"fmt"
	"io/ioutil"
	// "html"
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

func getGroupID(spreadsheet spreadsheet.Spreadsheet, course, secret string) (bool, string) {
	mainsheet, err := spreadsheet.SheetByTitle("course")
	if err != nil {
        log.Println(err)
		return false, ""
	}
	for _, row := range mainsheet.Rows {
		if row[0].Value == course {
			return row[1].Value == secret, row[2].Value
		}
	}
	return false, ""
}

func SheetService(clientsecret string) (s *spreadsheet.Service, err error) {
	data, err := ioutil.ReadFile(clientsecret)
	if err != nil {
		return
	}

	conf, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return
	}

	s = spreadsheet.NewServiceWithClient(conf.Client(oauth2.NoContext))
	return
}


func CampusWire(w http.ResponseWriter, r *http.Request) {
	dotenv := "./serverless_function_source_code/.env"
	if _, err := os.Stat(dotenv); os.IsNotExist(err) {
		dotenv = ".env"
	}
	err := godotenv.Load(dotenv)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	spreadsheetID := os.Getenv("SHEETID")
	cwtoken := os.Getenv("CWTOKEN")

	params := r.URL.Query()
	course := params.Get("course")
	secret := params.Get("secret")
	if course == "" || secret == "" {
		http.Error(w, "missing course or secret", http.StatusBadRequest)
		return
	}

	clientsecret := "./serverless_function_source_code/client_secret.json"
	if _, err := os.Stat(clientsecret); os.IsNotExist(err) {
		clientsecret = "client_secret.json"
	}
	service, err := SheetService(clientsecret)
	if err != nil {
        log.Println(err)
		http.Error(w, "cannot access service account", http.StatusInternalServerError)
		return
	}
	spreadsheet, err := service.FetchSpreadsheet(spreadsheetID)
	if err != nil {
        log.Println(err)
		http.Error(w, "cannot access spreadsheet", http.StatusInternalServerError)
		return
	}

	ok, groupID := getGroupID(spreadsheet, course, secret)
	if !ok {
		http.Error(w, "invalid secret", http.StatusForbidden)
		return
	}

	records, err := getReputations(cwtoken, groupID)
	if err != nil {
        log.Println(err)
		http.Error(w, "fail to get reputation", http.StatusInternalServerError)
		return
	}

	coursesheet, err := spreadsheet.SheetByTitle(course)
	if err != nil {
        log.Println(err)
		http.Error(w, "fail to get course sheet", http.StatusInternalServerError)
		return
	}

	n := len(coursesheet.Rows)
	m := len(records[0])

	for i, record := range records {
		if i == 0 {
			continue
		}
		for j, field := range record {
			coursesheet.Update(n + i - 1, j, field)
		}
		coursesheet.Update(n + i - 1, m, time.Now().Format(time.RFC3339))
	}
	err = coursesheet.Synchronize()
	if err != nil {
        log.Println(err)
	}
}
