package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func SendSaveBillMessage(bill *models.Bill, snsClient svc.SNSType) error {
	billJson, _ := json.Marshal(bill)
	return snsClient.Publish(string(billJson), os.Getenv("SNS_TOPIC_ARN"), "save_bill")
}

func SendTweetMessage(text string, params *twitter.StatusUpdateParams, snsClient svc.SNSType) error {
	data := svc.TweetData{Text: text, Params: *params}
	tweetJson, _ := json.Marshal(data)
	return snsClient.Publish(string(tweetJson), os.Getenv("SNS_TOPIC_ARN"), "post_tweet")
}

func HandleTweet(bill *models.Bill, db *gorm.DB, snsClient svc.SNSType) error {
	var billForTweet models.Bill

	if !db.Where(&models.Bill{TweetID: bill.TweetID}).Take(&billForTweet).RecordNotFound() {
		// Duplicate record already handled, exit
		return nil
	}

	if bill.BillID == "" {
		bill.Active = false
		_ = SendSaveBillMessage(bill, snsClient)
		_ = SendTweetMessage(
			"Couldn't parse a bill identifier from the tweet",
			&twitter.StatusUpdateParams{InReplyToStatusID: *bill.TweetID},
			snsClient,
		)
		return nil
	}

	var existingBill models.Bill
	if db.Where(&models.Bill{BillID: bill.BillID}).Take(&existingBill).RecordNotFound() {
		ocdBill := bill.GetOCDBill()
		if ocdBill.ID == "" {
			// Tweet that a valid bill wasn't found
			bill.Active = false
			_ = SendSaveBillMessage(bill, snsClient)
			_ = SendTweetMessage(
				"Valid bill not found",
				&twitter.StatusUpdateParams{InReplyToStatusID: *bill.TweetID},
				snsClient,
			)
			return nil
		}
		// Tweet that the new bill is now being tracked, save
		_ = SendSaveBillMessage(bill, snsClient)
		_ = SendTweetMessage(
			fmt.Sprintf("Bill now being tracked, you can follow with #%s", bill.BillID),
			&twitter.StatusUpdateParams{InReplyToStatusID: *bill.TweetID},
			snsClient,
		)
	} else {
		// Tweet standard reply about already being able to follow it with hashtag
		existingBill.LastTweetID = bill.LastTweetID
		_ = SendSaveBillMessage(&existingBill, snsClient)
		_ = SendTweetMessage(
			fmt.Sprintf("Bill now being tracked, you can follow with #%s", bill.BillID),
			&twitter.StatusUpdateParams{InReplyToStatusID: *bill.TweetID},
			snsClient,
		)
	}
	return nil
}

func handler(request events.SNSEvent) error {
	if len(request.Records) < 0 {
		return nil
	}
	message := request.Records[0].SNS.Message
	db, err := gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	snsClient := svc.NewSNSClient()
	if err != nil {
		// Re-publish message if DB error to retry
		snsClient.Publish(message, os.Getenv("SNS_TOPIC_ARN"), "tweets")
		panic(err)
	}
	defer db.Close()

	var bill models.Bill
	err = json.Unmarshal([]byte(message), &bill)
	if err != nil {
		panic(err)
	}
	return HandleTweet(&bill, db, snsClient)
}

func main() {
	lambda.Start(handler)
}
