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
	if bill.ParseBillID("@chibillbot test O2019-5305") != "O20195305" {
		t.Errorf("ParseBillID should parse 'O2019-5305' correctly")
	}
	if bill.ParseBillID("@chibillbot 02019-5305") != "" {
		t.Errorf("ParseBillID should return an empty string for '02019-5305")
	}
}

func TestGetCleanBillID(t *testing.T) {
	var apiBillId string
	bill := Bill{BillID: "O20151111"}
	apiBillId = bill.GetCleanBillID()
	if apiBillId != "O2015-1111" {
		t.Errorf("GetCleanBillID should return 'O2015-1111', got %s", apiBillId)
	}
	bill.BillID = "FL20101"
	apiBillId = bill.GetCleanBillID()
	if apiBillId != "FL2010-1" {
		t.Errorf("GetCleanBillID should return 'FL2010-1', got %s", apiBillId)
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
		Title: "Testing bill",
		Classification: "Ordinance",
		URL:            "https://chicago.legistar.com",
		BillID:         "O201011",
		Data:           `[]`,
	}
	tweetEnd := "See more at https://chicago.legistar.com #O201011"
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Testing bill. %s", tweetEnd) {
		t.Errorf("Tweet with no actions is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `[{"action": "fake"}]`
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Testing bill. %s", tweetEnd) {
		t.Errorf("Tweet with invalid action is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `[{"action": "Introduced", "actor": "Chicago City Council"}]`
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Testing bill was introduced in Chicago City Council. %s", tweetEnd) {
		t.Errorf("Tweet for introduction is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `[{"action": "Referred"}]`
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Testing bill was referred to committee. %s", tweetEnd) {
		t.Errorf("Tweet for referral with no entity is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `[{"action": "Referred", "actor": "", "committee": "Test Committee"}]`
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Testing bill was referred to the Test Committee. %s", tweetEnd) {
		t.Errorf("Tweet for referral with committee is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `[{"action": "Recommended for Passage", "actor": "Test Committee"}]`
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Testing bill was recommended to pass by the Test Committee. %s", tweetEnd) {
		t.Errorf("Tweet for committee passage is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `[{"action": "Passed"}]`
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Testing bill passed. %s", tweetEnd) {
		t.Errorf("Tweet for passage is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `[{"action": "Passed"}, {"action": "Referred"}]`
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Testing bill passed. %s", tweetEnd) {
		t.Errorf("Tweet for most recent action is incorrect: %s", bill.CreateTweet())
	}
	bill.Title = ""
	if bill.CreateTweet() != fmt.Sprintf("O2010-11 passed. %s", tweetEnd) {
		t.Errorf("Tweet for bill without title is incorrect: %s", bill.CreateTweet())
	}
	bill.Data = `[{"action": "Recommended for Passage", "actor": "Committee on Ethics and Government Oversight"}]`
	bill.Title = "Certification of city funding requirement for Laborers' and Retirement Board Employees Annuity and Benefit Fund of Chicago for tax year 2020, payment year 2021"
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Certification of city funding requirement for Laborers' and Retirement Board Employees Annuity and Benefit Fund of Chicago for tax year 2020, pay... was recommended to pass by the Committee on Ethics and Government Oversight. %s", tweetEnd) {
		t.Errorf("Clipped title tweet text with action is incorrect: %s", bill.CreateTweet())
	}
	bill.Title = "Certification of city funding requirement for Laborers' and Retirement Board Employees Annuity and Benefit Fund of Chicago for tax year 2020, payment year 2021                                                                  "
	bill.Data = `[]`
	if bill.CreateTweet() != fmt.Sprintf("O2010-11: Certification of city funding requirement for Laborers' and Retirement Board Employees Annuity and Benefit Fund of Chicago for tax year 2020, payment year 2021... %s", tweetEnd) {
		t.Errorf("Clipped title tweet text with no actions is incorrect: %s", bill.CreateTweet())
	}
}
