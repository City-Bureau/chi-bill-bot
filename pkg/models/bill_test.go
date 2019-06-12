package models

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestUnmarshalBill(t *testing.T) {
	var bill Bill
	billJson := `{"tweet_id": 10, "tweet_text": "R2019-55", "last_tweet_id": 1}`
	err := json.Unmarshal([]byte(billJson), &bill)
	if err != nil || *bill.TweetID != 10 || *bill.LastTweetID != 1 {
		t.Errorf("Should correctly unmarshal")
	}
}

func TestParseBillID(t *testing.T) {
	bill := Bill{}
	if bill.ParseBillID("o 2015 1111") != "O20151111" {
		t.Errorf("ParseBillID should correctly parse 'o 2015 1111'")
	}
	if bill.ParseBillID("o 2015") != "" {
		t.Errorf("ParseBillID should return an empty string for 'o 2015'")
	}
	if bill.ParseBillID("O-2015-12") != "O201512" {
		t.Errorf("ParseBillID should handle hyphens in 'O-2015-12'")
	}
	if bill.ParseBillID("@chicagoledger O2018-7001 test") != "O20187001" {
		t.Errorf("ParseBillID should parse 'O2018-7001' correctly")
	}
}

func TestGetAPIBillID(t *testing.T) {
	var apiBillId string
	bill := Bill{BillID: "O20151111"}
	apiBillId = bill.GetAPIBillID()
	if apiBillId != "O2015-1111" {
		t.Errorf("GetAPIBillID should return 'O2015-1111', got %s", apiBillId)
	}
	bill.BillID = "FL20101"
	apiBillId = bill.GetAPIBillID()
	if apiBillId != "FL2010-1" {
		t.Errorf("GetAPIBillID should return 'FL2010-1', got %s", apiBillId)
	}
}

func TestSetNextRun(t *testing.T) {
	bill := Bill{}
	bill.SetNextRun()
	if bill.NextRun.Hour() < 9 || bill.NextRun.Hour() > 17 {
		t.Errorf("Hour: %d is outside range 9AM-10PM", bill.NextRun.Hour())
	}
}

func TestCreateTweet(t *testing.T) {
	bill := Bill{
		BillID: "O201011",
		Data:   `{"extras": {"local_classification": "Ordinance"}, "actions": []}`,
	}
	tweetEnd := "See more at https://chicago.councilmatic.org/legislation/o2010-11 #O201011"
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11. %s", tweetEnd) {
		t.Errorf("Tweet with no actions is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {"local_classification": "Ordinance"}, "actions": [{"classification": ["fake"]}]}`
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11. %s", tweetEnd) {
		t.Errorf("Tweet with invalid action is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {"local_classification": "Ordinance"}, "actions": [{"classification": ["introduction"], "organization": {"name": "Chicago City Council"}}]}`
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11 was introduced in Chicago City Council. %s", tweetEnd) {
		t.Errorf("Tweet for introduction is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {"local_classification": "Ordinance"}, "actions": [{"classification": ["committee-referral"]}]}`
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11 was referred to committee. %s", tweetEnd) {
		t.Errorf("Tweet for referral with no entity is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {"local_classification": "Ordinance"}, "actions": [{"classification": ["referral-committee"], "related_entities": [{"name": "Test Committee"}]}]}`
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11 was referred to the Test Committee. %s", tweetEnd) {
		t.Errorf("Tweet for referral with entity is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {"local_classification": "Ordinance"}, "actions": [{"classification": ["committee-passage-favorable"], "organization": {"name": "Test Committee"}}]}`
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11 was recommended to pass by the Test Committee. %s", tweetEnd) {
		t.Errorf("Tweet for committee passage is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {"local_classification": "Ordinance"}, "actions": [{"classification": ["amendment-passage"], "organization": {"name": "Test Committee"}}]}`
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11 was amended in the Test Committee. %s", tweetEnd) {
		t.Errorf("Tweet for amendment passage is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {"local_classification": "Ordinance"}, "actions": [{"classification": ["passage"]}]}`
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11 passed. %s", tweetEnd) {
		t.Errorf("Tweet for passage is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {"local_classification": "Ordinance"}, "actions": [{"classification": ["executive-signature"]}]}`
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11 was signed by the mayor. %s", tweetEnd) {
		t.Errorf("Tweet for executive signature is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {"local_classification": "Ordinance"}, "actions": [{"classification": ["passage"]}, {"classification": ["executive-signature"]}]}`
	if bill.CreateTweet() != fmt.Sprintf("Ordinance O2010-11 was signed by the mayor. %s", tweetEnd) {
		t.Errorf("Tweet for last action is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `{"extras": {}, "actions": [{"classification": ["passage"]}]}`
	if bill.CreateTweet() != fmt.Sprintf("O2010-11 passed. %s", tweetEnd) {
		t.Errorf("Tweet missing classification is incorrect: %s", bill.CreateTweet())
	}
}
