package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"
	"github.com/getsentry/sentry-go"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func updateBill(bill models.Bill, actions []models.LegistarAction, snsClient svc.SNSType) error {
	billActions := bill.GetActions()

	// If the bill is new or has changed tweet it out
	if bill.NextRun == nil || len(actions) > len(billActions) {
		actionJSON, err := json.Marshal(actions)
		if err != nil {
			return err
		}
		bill.Data = string(actionJSON)
		billURL := bill.GetTweetURL()
		data := svc.TweetData{Text: bill.CreateTweet(billURL)}
		tweetJSON, err := json.Marshal(data)
		if err != nil {
			return err
		}
		err = snsClient.Publish(string(tweetJSON), os.Getenv("SNS_TOPIC_ARN"), "post_tweet")
		if err != nil {
			return err
		}
	}
	bill.SetNextRun()
	billJSON, err := json.Marshal(bill)
	if err != nil {
		return err
	}
	// Return potential errors from saving last, because if the tweet failed then it will
	// still be retried if there's a difference from what's in the database
	return snsClient.Publish(string(billJSON), os.Getenv("SNS_TOPIC_ARN"), "save_bill")
}

func handler(request events.SNSEvent) error {
	if len(request.Records) < 0 {
		return nil
	}
	message := request.Records[0].SNS.Message
	snsClient := svc.NewSNSClient()

	var bill models.Bill
	err := json.Unmarshal([]byte(message), &bill)
	// Log errors because we don't want to trigger Lambda's retries
	if err != nil {
		_ = snsClient.Publish(message, os.Getenv("SNS_TOPIC_ARN"), "update_bill")
		sentry.CaptureException(err)
		log.Println(err)
		return nil
	}

	// Get new data for bill, check if it's changed
	title, cls, actions, err := bill.FetchBillData()
	if err != nil {
		sentry.CaptureException(err)
		return err
	}
	bill.Title = title
	bill.Classification = cls

	err = updateBill(bill, actions, snsClient)
	// Only log this error since it just prevented
	if err != nil {
		_ = snsClient.Publish(message, os.Getenv("SNS_TOPIC_ARN"), "update_bill")
		sentry.CaptureException(err)
		log.Println(err)
	}
	return nil
}

func main() {
	_ = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
		Transport: &sentry.HTTPSyncTransport{
			Timeout: 5 * time.Second,
		},
	})

	lambda.Start(handler)
}
