package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
)

func QueryMentions(sinceTweetId string, twttr svc.Twitter, snsClient svc.SNSType) error {
	var params twitter.MentionTimelineParams
	if sinceTweetId != "" {
		sinceTweetIdInt, err := strconv.ParseInt(sinceTweetId, 10, 64)
		if err != nil {
			return err
		}
		params = twitter.MentionTimelineParams{SinceID: sinceTweetIdInt}
	}
	tweets, err := twttr.GetMentions(&params)

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
		tweetBill := &models.Bill{
			TweetID:     &tweet.ID,
			TweetText:   tweet.FullText,
			LastTweetID: &lastTweetId,
		}

		// Load bill data from tweet
		tweetBill.BillID = tweetBill.ParseBillID(tweetBill.TweetText)
		if tweetBill.BillID != "" {
			billData, _ := tweetBill.LoadBillData()
			billJson, _ := json.Marshal(billData)
			tweetBill.Data = string(billJson)
		}
		tweetBillJson, _ := json.Marshal(tweetBill)
		err = snsClient.Publish(string(tweetBillJson), os.Getenv("SNS_TOPIC_ARN"), "handle_tweet")
		if err != nil {
			return err
		}
	}
	return nil
}

func handler(request events.SNSEvent) error {
	err := QueryMentions(request.Records[0].SNS.Message, svc.NewTwitterClient(), svc.NewSNSClient())
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func main() {
	lambda.Start(handler)
}
