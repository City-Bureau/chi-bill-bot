package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func QueryMentions(db *gorm.DB, twttr svc.Twitter, snsClient svc.SNSType) error {
	var bill models.Bill

	db.Order("last_tweet_id desc").First(&bill)
	tweets, err := twttr.GetMentions(&twitter.MentionTimelineParams{SinceID: bill.LastTweetID.Int64})

	if err != nil {
		log.Fatal(err)
	}

	// Get the last tweet in the list of tweets, assign that to all
	lastTweetId := tweets[len(tweets)-1].ID

	// Iterate through mentions, publishing each to SNS topic
	for _, tweet := range tweets {
		tweetBill := &models.Bill{
			TweetID:     sql.NullInt64{Int64: tweet.ID, Valid: true},
			TweetText:   tweet.FullText,
			LastTweetID: sql.NullInt64{Int64: lastTweetId, Valid: true},
		}
		tweetBillJson, _ := json.Marshal(tweetBill)

		err = snsClient.Publish(string(tweetBillJson), os.Getenv("SNS_TOPIC_ARN"), "tweets")
	}
	return nil
}

func handler(request events.CloudWatchEvent) error {
	db, err := gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	return QueryMentions(db, svc.NewTwitterClient(), svc.NewSNSClient())
}

func main() {
	lambda.Start(handler)
}
