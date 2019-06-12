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
	Date            string   `json:"date"`
	Description     string   `json:"description"`
	Classification  []string `json:"classification"`
	RelatedEntities []Entity `json:"related_entities"`
	Organization    Organization
}

type Extras struct {
	Classification string `json:"local_classification,omitempty"`
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
	Extras         Extras
}

type OCDResponse struct {
	Results []OCDBill `json:"results"`
}

type Bill struct {
	PK          uint   `gorm:"primary_key"`
	TweetID     *int64 `json:"tweet_id,omitempty"`
	TweetUser   string `gorm:"size:250" json:"tweet_user"`
	TweetText   string `gorm:"size:300" json:"tweet_text"`
	LastTweetID *int64 `json:"last_tweet_id,omitempty"`
	BillID      string `gorm:"size:25" json:"id,omitempty"`
	Active      bool   `gorm:"default:true"`
	Data        string `gorm:"type:text"`
	NextRun     *time.Time
}

func GetOCDRes(url string) ([]byte, error) {
	client := http.Client{
		Timeout: time.Second * 180,
	}

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
	billId := b.GetAPIBillID()
	ocdBill := b.GetOCDBill()
	billCls := ocdBill.Extras.Classification
	if billCls != "" {
		billCls = fmt.Sprintf("%s ", billCls)
	}
	url := fmt.Sprintf("https://chicago.councilmatic.org/legislation/%s", strings.ToLower(billId))
	if len(ocdBill.Actions) == 0 {
		return fmt.Sprintf("%s%s. See more at %s #%s", billCls, billId, url, b.BillID)
	}
	action := ocdBill.Actions[len(ocdBill.Actions)-1]

	actionText := fmt.Sprintf("%s%s", billCls, billId)
	classification := ""
	if len(action.Classification) > 0 {
		classification = action.Classification[0]
	}
	switch cls := classification; cls {
	case "introduction":
		actionText = fmt.Sprintf("%s%s was introduced in %s", billCls, billId, action.Organization.Name)
	case "filing":
		actionText = fmt.Sprintf("%s%s was placed on file", billCls, billId)
	case "committee-referral", "referral-committee":
		if len(action.RelatedEntities) > 0 {
			actionText = fmt.Sprintf("%s%s was referred to the %s", billCls, billId, action.RelatedEntities[0].Name)
		} else {
			actionText = fmt.Sprintf("%s%s was referred to committee", billCls, billId)
		}
	case "committee-passage-favorable":
		actionText = fmt.Sprintf("%s%s was recommended to pass by the %s", billCls, billId, action.Organization.Name)
	case "amendment-passage":
		actionText = fmt.Sprintf("%s%s was amended in the %s", billCls, billId, action.Organization.Name)
	case "passage":
		actionText = fmt.Sprintf("%s%s passed", billCls, billId)
	case "executive-signature":
		actionText = fmt.Sprintf("%s%s was signed by the mayor", billCls, billId)
	}

	return fmt.Sprintf("%s. See more at %s #%s", actionText, url, b.BillID)
}

func (b *Bill) SetNextRun() {
	// Set NextRun to a time in the future within a defined range
	loc, _ := time.LoadLocation("America/Chicago")
	now := time.Now().In(loc)
	// Make sure it's between 9AM and 10PM Chicago
	diffHours := 24
	if now.Hour() < 9 {
		diffHours = diffHours + (9 - now.Hour())
	} else if now.Hour() > 17 {
		diffHours = diffHours - (now.Hour() - 17)
	}
	nextRun := now.Add(time.Hour * time.Duration(diffHours))
	b.NextRun = &nextRun
}
