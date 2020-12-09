package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/getsentry/sentry-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func saveBillAndTweet(text string, bill *models.Bill, snsClient svc.SNSType) error {
	billJSON, err := json.Marshal(bill)
	if err != nil {
		return err
	}
	data := svc.TweetData{
		Text:   fmt.Sprintf("@%s %s", bill.TweetUser, text),
		Params: twitter.StatusUpdateParams{InReplyToStatusID: *bill.TweetID},
	}
	tweetJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = snsClient.Publish(string(tweetJSON), os.Getenv("SNS_TOPIC_ARN"), "post_tweet")
	if err != nil {
		return err
	}
	return snsClient.Publish(string(billJSON), os.Getenv("SNS_TOPIC_ARN"), "update_bill")
}

func handleTweet(bill *models.Bill, db *gorm.DB, snsClient svc.SNSType) error {
	var billForTweet models.Bill

	if !db.Where(&models.Bill{TweetID: bill.TweetID}).Take(&billForTweet).RecordNotFound() {
		// Duplicate record already handled, exit
		return nil
	}

	if bill.BillID == "" {
		bill.Active = false
		// Don't post a tweet for unmatched bill, otherwise will go for all tweets with mentions
		billJSON, _ := json.Marshal(bill)
		return snsClient.Publish(string(billJSON), os.Getenv("SNS_TOPIC_ARN"), "save_bill")
	}

	var existingBill models.Bill
	if db.Where(&models.Bill{BillID: bill.BillID}).Take(&existingBill).RecordNotFound() {
		billCls := bill.Classification
		if billCls != "" {
			billCls = fmt.Sprintf("%s ", billCls)
		}
		if bill.URL == "" {
			// Tweet that a valid bill wasn't found
			bill.Active = false
			return saveBillAndTweet(
				"We couldn't find a Chicago City Council bill with that ID",
				bill,
				snsClient,
			)
		}
		// Tweet that the new bill is now being tracked, save
		return saveBillAndTweet(
			fmt.Sprintf(
				"We're now tracking Chicago City Council %s%s. You can follow along with #%s—we'll tweet when this legislation moves. %s",
				billCls,
				bill.GetCleanBillID(),
				bill.BillID,
				bill.URL,
			),
			bill,
			snsClient,
		)
	}

	// Tweet standard reply about already being able to follow it with hashtag
	existingBill.LastTweetID = bill.LastTweetID
	billCls := bill.Classification
	if billCls != "" {
		billCls = fmt.Sprintf("%s ", billCls)
	}
	return saveBillAndTweet(
		fmt.Sprintf(
			"We're already tracking %s%s. You can follow along with #%s—we'll tweet when this legislation moves. %s",
			billCls,
			bill.GetCleanBillID(),
			existingBill.BillID,
			bill.URL,
		),
		&existingBill,
		snsClient,
	)
}

func handler(request events.SNSEvent) error {
	if len(request.Records) < 0 {
		return nil
	}
	message := request.Records[0].SNS.Message
	db, err := gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?parseTime=true",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	snsClient := svc.NewSNSClient()
	if err != nil {
		sentry.CaptureException(err)
		// Log failure to trigger Lambda retry
		log.Fatal(err)
		return err
	}
	defer db.Close()

	var bill models.Bill
	err = json.Unmarshal([]byte(message), &bill)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
		return err
	}
	err = handleTweet(&bill, db, snsClient)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
		return err
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
