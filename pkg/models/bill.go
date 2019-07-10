package models

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type LegistarAction struct {
	Date      time.Time `json:"date,omitempty"`
	Actor     string    `json:"actor,omitempty"`
	Action    string    `json:"action,omitempty"`
	Committee string    `json:"committee,omitempty"`
}

type Bill struct {
	PK             uint   `gorm:"primary_key"`
	TweetID        *int64 `json:"tweet_id,omitempty"`
	TweetUser      string `gorm:"size:250" json:"tweet_user"`
	TweetText      string `gorm:"size:300" json:"tweet_text"`
	LastTweetID    *int64 `json:"last_tweet_id,omitempty"`
	BillID         string `gorm:"size:25" json:"id,omitempty"`
	Classification string `gorm:"size:250" json:"classification"`
	URL            string `gorm:"size:250" json:"url"`
	Active         bool   `gorm:"default:true"`
	Data           string `gorm:"type:text"`
	NextRun        *time.Time
}

func (b *Bill) ParseBillID(text string) string {
	billRe := regexp.MustCompile(`(^| )[a-zA-Z]{1,4}[\-\s]*\d{4}[\-\s]*\d{1,5}`)
	spacerRe := regexp.MustCompile(`[\s-]+`)
	billText := strings.Trim(billRe.FindString(text), " ")
	billUpper := strings.ToUpper(spacerRe.ReplaceAllLiteralString(billText, ""))
	// Special exception for Orders
	return strings.Replace(billUpper, "OR", "Or", 1)
}

func (b *Bill) GetCleanBillID() string {
	// Return bill ID in format for API
	billRe := regexp.MustCompile(`(?P<type>[A-Za-z]+)(?P<year>\d{4})(?P<id>\d+)`)
	billMatch := billRe.FindStringSubmatch(b.BillID)
	result := make(map[string]string)
	for i, name := range billRe.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = billMatch[i]
		}
	}
	return fmt.Sprintf("%s%s-%s", result["type"], result["year"], result["id"])
}

func (b *Bill) GetActions() []LegistarAction {
	var actions []LegistarAction

	err := json.Unmarshal([]byte(b.Data), &actions)
	if err != nil {
		log.Fatal(err)
	}
	return actions
}

func (b *Bill) SearchBill() (string, error) {
	response, err := http.Get("https://chicago.legistar.com/Legislation.aspx")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", err
	}

	viewstate, _ := document.Find("input[name='__VIEWSTATE']").First().Attr("value")
	eventvalidation, _ := document.Find("input[name='__EVENTVALIDATION']").First().Attr("value")
	payload := url.Values{
		"__EVENTARGUMENT":                                        {""},
		"__VIEWSTATE":                                            {viewstate},
		"__EVENTVALIDATION":                                      {eventvalidation},
		"ctl00$ContentPlaceHolder1$txtFil":                       {b.GetCleanBillID()},
		"ctl00_ContentPlaceHolder1_lstMax_ClientState":           {"{\"value\":\"1000000\"}"},
		"ctl00_ContentPlaceHolder1_lstYearsAdvanced_ClientState": {"{\"value\":\"All\"}"},
		"ctl00$ContentPlaceHolder1$btnSearch":                    {"Search Legislation"},
	}
	response, err = http.PostForm("https://chicago.legistar.com/Legislation.aspx", payload)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	document, err = goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", err
	}

	var billUrl string
	document.Find(".rgMasterTable tbody tr > td:first-child a").Each(func(index int, element *goquery.Selection) {
		if element != nil && strings.Trim(element.Text(), " ") == b.GetCleanBillID() {
			billUrl, _ = element.Attr("href")
		}
	})

	return fmt.Sprintf("https://chicago.legistar.com/%s", billUrl), nil
}

func (b *Bill) FetchBillData() (string, []LegistarAction, error) {
	var actions []LegistarAction

	response, err := http.Get(b.URL)
	if err != nil {
		return "", actions, err
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", actions, err
	}

	status := strings.Trim(document.Find("#ctl00_ContentPlaceHolder1_lblStatus2").First().Text(), "\n\t ")
	classification := strings.Trim(document.Find("#ctl00_ContentPlaceHolder1_lblType2").First().Text(), "\n\t ")
	committee := strings.Trim(document.Find("#ctl00_ContentPlaceHolder1_hypInControlOf2").First().Text(), "\n\t ")

	document.Find(".rgMasterTable tbody tr").Each(func(index int, element *goquery.Selection) {
		action := LegistarAction{}
		element.Find("td").Each(func(tdIdx int, tdEl *goquery.Selection) {
			if tdIdx == 0 {
				action.Date, _ = time.Parse("1/2/2006", strings.Trim(tdEl.Text(), "\n\t "))
			} else if tdIdx == 2 {
				action.Actor = strings.Trim(tdEl.Text(), "\n\t ")
			} else if tdIdx == 3 {
				action.Action = strings.Trim(tdEl.Text(), "\n\t ")
				if action.Action == "" {
					action.Action = status
				}
			}
			if strings.Contains(action.Action, "Referred") {
				action.Committee = committee
			}
		})
		actions = append(actions, action)
	})
	return classification, actions, nil
}

func (b *Bill) CreateTweet() string {
	billId := b.GetCleanBillID()
	billCls := b.Classification
	actions := b.GetActions()
	if billCls != "" {
		billCls = fmt.Sprintf("%s ", billCls)
	}
	if len(actions) == 0 {
		return fmt.Sprintf("%s%s. See more at %s #%s", billCls, billId, b.URL, b.BillID)
	}
	// Pull the first action which is the most recent on Legistar
	action := actions[0]
	actionText := fmt.Sprintf("%s%s", billCls, billId)
	switch cls := action.Action; cls {
	case "Introduced", "Direct Introduction":
		actionText = fmt.Sprintf("%s%s was introduced in %s", billCls, billId, action.Actor)
	case "Placed on File":
		actionText = fmt.Sprintf("%s%s was placed on file", billCls, billId)
	case "Referred", "Re-Referred":
		if action.Committee != "" {
			actionText = fmt.Sprintf("%s%s was referred to the %s", billCls, billId, action.Committee)
		} else {
			actionText = fmt.Sprintf("%s%s was referred to committee", billCls, billId)
		}
	case "Recommended for Passage":
		actionText = fmt.Sprintf("%s%s was recommended to pass by the %s", billCls, billId, action.Actor)
	case "Recommended Do Not Pass":
		actionText = fmt.Sprintf("%s%s was recommended not to pass by the %s", billCls, billId, action.Actor)
	case "Recommended for Re-referral":
		actionText = fmt.Sprintf("%s%s was recommended for re-referral by the %s", billCls, billId, action.Actor)
	case "Passed", "Passed as Substitute":
		actionText = fmt.Sprintf("%s%s passed", billCls, billId)
	case "Failed to Pass":
		actionText = fmt.Sprintf("%s%s failed to pass", billCls, billId)
	case "Introduced (Agreed Calendar)", "Adopted":
		actionText = fmt.Sprintf("%s%s was adopted", billCls, billId)
	case "Approved", "Repealed", "Vetoed", "Tabled", "Withdrawn":
		actionText = fmt.Sprintf("%s%s was %s", billCls, billId, strings.ToLower(cls))
	}

	return fmt.Sprintf("%s. See more at %s #%s", actionText, b.URL, b.BillID)
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
