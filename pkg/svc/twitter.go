package svc

import (
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Twitter interface {
	PostTweet(string, *twitter.StatusUpdateParams) error
	GetMentions(*twitter.MentionTimelineParams) ([]twitter.Tweet, error)
}

type TwitterClient struct {
	Client *twitter.Client
}

type TweetData struct {
	Text   string                     `json:"text"`
	Params twitter.StatusUpdateParams `json:"params"`
}

func NewTwitterClient() *TwitterClient {
	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessSecret := os.Getenv("TWITTER_ACCESS_SECRET")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	return &TwitterClient{Client: twitter.NewClient(httpClient)}
}

func (t *TwitterClient) PostTweet(tweet string, params *twitter.StatusUpdateParams) error {
	// _, _, err := t.Client.Statuses.Update(tweet, params)
	// return err
	log.Printf(tweet)
	return nil
}

func (t *TwitterClient) GetMentions(params *twitter.MentionTimelineParams) ([]twitter.Tweet, error) {
	tweets, _, err := t.Client.Timelines.MentionTimeline(params)
	return tweets, err
}
