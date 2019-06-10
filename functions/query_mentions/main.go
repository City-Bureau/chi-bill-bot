package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
)

func QueryMentions(twttr svc.Twitter, snsClient svc.SNSType) error {
	tweets, err := twttr.GetMentions(&twitter.MentionTimelineParams{})
	if err != nil {
		return err
	}

	if len(tweets) == 0 {
		return nil
	}
	// Get the last tweet in the list of tweets, assign that to all
	lastTweetId := tweets[len(tweets)-1].ID

	// Iterate through mentions, publishing each to SNS topic
	for _, tweet := range tweets {
		// Ignore if more than 48 hours old
		createdTime, _ := tweet.CreatedAtTime()
		if createdTime.Before(time.Now().Add(time.Hour * -48)) {
			continue
		}
		// TODO: Figure out ExtendedTweet
		tweetBill := &models.Bill{
			TweetID:     &tweet.ID,
			TweetText:   tweet.Text,
			LastTweetID: &lastTweetId,
		}

		// Load bill data from tweet
		tweetBill.BillID = tweetBill.ParseBillID(tweetBill.TweetText)
		log.Println(tweet.ID)
		log.Println(tweetBill.BillID)
		if tweetBill.BillID == "" {
			continue
		}
		billData, _ := tweetBill.LoadBillData()
		billJson, _ := json.Marshal(billData)
		tweetBill.Data = string(billJson)
		tweetBillJson, _ := json.Marshal(tweetBill)
		err = snsClient.Publish(string(tweetBillJson), os.Getenv("SNS_TOPIC_ARN"), "handle_tweet")
		if err != nil {
			return err
		}
	}
	return nil
}

func handler(request events.CloudWatchEvent) error {
	err := QueryMentions(svc.NewTwitterClient(), svc.NewSNSClient())
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func main() {
	lambda.Start(handler)
}
