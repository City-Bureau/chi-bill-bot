package models

import (
	"encoding/json"
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
