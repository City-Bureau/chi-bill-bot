package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func HandleBill(tweet string, bill *models.Bill, db *gorm.DB, twttr svc.Twitter, snsClient svc.SNSType) error {
	// Tweet the updated bill
	twttr.PostTweet(tweet, &twitter.StatusUpdateParams{})

	nextRun := time.Now().Add(24 * time.Hour)
	bill.NextRun = &nextRun
	db.Save(&bill)

	return nil
}

func handler(request events.SNSEvent) error {
	if len(request.Records) < 0 {
		return nil
	}
	message := request.Records[0].SNS.Message
	db, err := gorm.Open("mysql", "CONN")
	snsClient := svc.NewSNSClient()
	if err != nil {
		// Re-publish message if DB error to retry
		snsClient.Publish(message, os.Getenv("SNS_TOPIC_ARN"))
		panic(err)
	}
	defer db.Close()

	var bill models.Bill
	err = json.Unmarshal([]byte(message), bill)
	if err != nil {
		return err
	}

	return HandleBill(bill.CreateTweet(), &bill, db, svc.NewTwitterClient(), snsClient)
}

func main() {
	lambda.Start(handler)
}
