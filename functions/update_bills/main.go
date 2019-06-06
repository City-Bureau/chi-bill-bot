package main

import (
	"encoding/json"
	"os"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func UpdateBills(bills []models.Bill, snsClient *svc.SNSClient) error {
	// Iterate through mentions, publishing each to SNS topic
	for _, bill := range bills {
		billData, _ := bill.GetOCDBillData()
		ocdBill := bill.GetOCDBill()
		billDataJson, _ := json.Marshal(billData)
		bill.SetNextRun()
		billJson, _ := json.Marshal(bill)
		// If the bill has changed, tweet it out
		if len(ocdBill.Actions) > len(billData.Actions) {
			bill.Data = string(billDataJson)
			data := svc.TweetData{Text: bill.CreateTweet()}
			tweetJson, _ := json.Marshal(data)
			snsClient.Publish(string(tweetJson), os.Getenv("SNS_TOPIC_ARN"), "post_tweet")
			// TODO: Add a check for dead bills with no activity in certain duration
		}
		snsClient.Publish(string(billJson), os.Getenv("SNS_TOPIC_ARN"), "save_bill")
	}
	return nil
}

func handler(request events.SNSEvent) error {
	if len(request.Records) < 0 {
		return nil
	}
	message := request.Records[0].SNS.Message

	var bills []models.Bill
	err := json.Unmarshal([]byte(message), &bills)
	if err != nil {
		return err
	}

	return UpdateBills(bills, svc.NewSNSClient())
}

func main() {
	lambda.Start(handler)
}
