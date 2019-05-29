package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Sponsorship struct {
	Type           string `json:"entity_type"`
	Name           string `json:"entity_name"`
	ID             string `json:"entity_id"`
	Classification string `json:"classification"`
	Primary        bool   `json:"primary"`
}

type Organization struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Entity struct {
	Type           string `json:"entity_type"`
	Name           string `json:"name"`
	PersonID       string `json:"person_id"`
	OrganizationID string `json:"organizaton_id"`
}

type Action struct {
	Date           string   `json:"date"`
	Description    string   `json:"description"`
	Classification []string `json:"classification"`
}

type OCDBill struct {
	ID             string        `json:"id"`
	Identifier     string        `json:"identifier"`
	CreatedAt      string        `json:"created_at"`
	UpdatedAt      string        `json:"updated_at"`
	Title          string        `json:"title"`
	Classification []string      `json:"classification"`
	Sponsorships   []Sponsorship `json:"sponsorships,omitempty"`
	Actions        []Action      `json:"actions,omitempty"`
	Entities       []Entity      `json:"related_entities,omitempty"`
}

type OCDResponse struct {
	Results []OCDBill `json:"results"`
}

type Bill struct {
	PK          uint   `gorm:"primary_key"`
	TweetID     *int64 `json:"tweet_id,omitempty"`
	TweetText   string `gorm:"size:300" json:"tweet_text"`
	LastTweetID *int64 `json:"last_tweet_id,omitempty"`
	BillID      string `gorm:"size:25" json:"id,omitempty"`
	Active      bool   `gorm:"default:true"`
	Data        string `gorm:"type:text"`
	NextRun     *time.Time
}

func GetOCDRes(url string) ([]byte, error) {
	client := http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	return body, nil
}

func (b *Bill) ParseBillID(text string) string {
	billRe := regexp.MustCompile(`[a-zA-Z]{1,4}[\-\s]*\d{4}[\-\s]*\d{1,5}`)
	spacerRe := regexp.MustCompile(`[\s-]+`)
	billText := billRe.FindString(text)
	return strings.ToUpper(spacerRe.ReplaceAllLiteralString(billText, ""))
}

func (b *Bill) LoadBillData() (OCDBill, error) {
	apiBillId := b.GetAPIBillID()
	billsUrl := fmt.Sprintf("https://ocd.datamade.us/bills/?identifier=%s", apiBillId)
	billsRes, _ := GetOCDRes(billsUrl)
	var billsOcdRes OCDResponse
	jsonErr := json.Unmarshal(billsRes, &billsOcdRes)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	if len(billsOcdRes.Results) == 0 {
		return OCDBill{}, nil
	}

	billUrl := fmt.Sprintf("https://ocd.datamade.us/%s", billsOcdRes.Results[0].ID)
	billRes, _ := GetOCDRes(billUrl)
	var billOcd OCDBill
	billJsonErr := json.Unmarshal(billRes, &billOcd)
	if billJsonErr != nil {
		log.Fatal(billJsonErr)
	}

	return billOcd, nil
}

func (b *Bill) GetOCDBillData() (OCDBill, error) {
	var billOcd OCDBill

	initBillOcd := b.GetOCDBill()
	url := fmt.Sprintf("https://ocd.datamade.us/%s", initBillOcd.ID)
	ocdRes, _ := GetOCDRes(url)

	jsonErr := json.Unmarshal(ocdRes, &billOcd)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return billOcd, nil
}

func (b *Bill) GetAPIBillID() string {
	// Return bill ID in format for API
	billRe := regexp.MustCompile(`(?P<type>[A-Z]+)(?P<year>\d{4})(?P<id>\d+)`)
	billMatch := billRe.FindStringSubmatch(b.BillID)
	result := make(map[string]string)
	for i, name := range billRe.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = billMatch[i]
		}
	}
	return fmt.Sprintf("%s%s-%s", result["type"], result["year"], result["id"])
}

func (b *Bill) GetOCDBill() OCDBill {
	var billData OCDBill

	err := json.Unmarshal([]byte(b.Data), &billData)
	if err != nil {
		log.Fatal(err)
	}
	return billData
}

func (b *Bill) CreateTweet() string {
	billData := b.GetOCDBill()
	return fmt.Sprintf("Tweet about new bill titled: %s #%s", billData.Title, b.BillID)
}

func (b *Bill) SetNextRun() {
	// Set NextRun to a time in the future within a defined range
	nextRun := time.Now().Add(time.Hour * 24)
	b.NextRun = &nextRun
}
