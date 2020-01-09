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

func queryMentions(twttr svc.Twitter, snsClient svc.SNSType) error {
	tweets, err := twttr.GetMentions(&twitter.MentionTimelineParams{})
	if err != nil {
		return err
	}

	if len(tweets) == 0 {
		return nil
	}
	// Get the last tweet in the list of tweets, assign that to all
	lastTweetID := tweets[len(tweets)-1].ID

	// Iterate through mentions, publishing each to SNS topic
	for _, tweet := range tweets {
		// Ignore if more than 2 hours old or doesn't have a user
		createdTime, _ := tweet.CreatedAtTime()
		if createdTime.Before(time.Now().Add(time.Hour*-2)) || tweet.User == nil {
			continue
		}
		// TODO: Figure out ExtendedTweet
		tweetBill := &models.Bill{
			TweetID:     &tweet.ID,
			TweetText:   tweet.Text,
			TweetUser:   tweet.User.ScreenName,
			LastTweetID: &lastTweetID,
		}

		// Load bill data from tweet
		tweetBill.BillID = tweetBill.ParseBillID(tweetBill.TweetText)
		log.Println(tweet.ID)
		log.Println(tweetBill.BillID)
		if tweetBill.BillID == "" {
			continue
		}
		billURL, _ := tweetBill.SearchBill()
		tweetBill.URL = billURL
		title, cls, actions, _ := tweetBill.FetchBillData()
		tweetBill.Title = title
		tweetBill.Classification = cls
		actionJSON, _ := json.Marshal(actions)
		tweetBill.Data = string(actionJSON)
		tweetBillJSON, _ := json.Marshal(tweetBill)
		err = snsClient.Publish(string(tweetBillJSON), os.Getenv("SNS_TOPIC_ARN"), "handle_tweet")
		if err != nil {
			return err
		}
	}
	return nil
}

func handler(request events.CloudWatchEvent) error {
	err := queryMentions(svc.NewTwitterClient(), svc.NewSNSClient())
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func main() {
	lambda.Start(handler)
}
