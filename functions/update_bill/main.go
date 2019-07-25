package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func UpdateBill(bill models.Bill, actions []models.LegistarAction, snsClient svc.SNSType) error {
	billActions := bill.GetActions()

	// If the bill is new or has changed tweet it out
	if bill.NextRun == nil || len(actions) > len(billActions) {
		actionJson, err := json.Marshal(actions)
		if err != nil {
			return err
		}
		bill.Data = string(actionJson)
		billUrl := bill.GetTweetURL()
		data := svc.TweetData{Text: bill.CreateTweet(billUrl)}
		tweetJson, err := json.Marshal(data)
		if err != nil {
			return err
		}
		err = snsClient.Publish(string(tweetJson), os.Getenv("SNS_TOPIC_ARN"), "post_tweet")
		if err != nil {
			return err
		}
	}
	bill.SetNextRun()
	billJson, err := json.Marshal(bill)
	if err != nil {
		return err
	}
	// Return potential errors from saving last, because if the tweet failed then it will
	// still be retried if there's a difference from what's in the database
	return snsClient.Publish(string(billJson), os.Getenv("SNS_TOPIC_ARN"), "save_bill")
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
		snsClient.Publish(message, os.Getenv("SNS_TOPIC_ARN"), "update_bill")
		log.Println(err)
		return nil
	}

	// Get new data for bill, check if it's changed
	title, cls, actions, err := bill.FetchBillData()
	if err != nil {
		return err
	}
	bill.Title = title
	bill.Classification = cls

	err = UpdateBill(bill, actions, snsClient)
	// Only log this error since it just prevented
	if err != nil {
		snsClient.Publish(message, os.Getenv("SNS_TOPIC_ARN"), "update_bill")
		log.Println(err)
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
