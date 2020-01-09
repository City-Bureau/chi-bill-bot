package svc

import (
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// Twitter is an interface for TwitterClient and its associated mock
type Twitter interface {
	PostTweet(string, *twitter.StatusUpdateParams) error
	GetMentions(*twitter.MentionTimelineParams) ([]twitter.Tweet, error)
}

// TwitterClient implements the Twitter interface for working with the Twitter API
type TwitterClient struct {
	Client *twitter.Client
}

// TweetData is a simplified struct for working with tweets
type TweetData struct {
	Text   string                     `json:"text"`
	Params twitter.StatusUpdateParams `json:"params"`
}

// NewTwitterClient creates a TwitterClient from environment variable credentials
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

// PostTweet posts a tweet to Twitter
func (t *TwitterClient) PostTweet(tweet string, params *twitter.StatusUpdateParams) error {
	log.Printf(tweet)
	_, _, err := t.Client.Statuses.Update(tweet, params)
	return err
}

// GetMentions queries recent mentions on Twitter
func (t *TwitterClient) GetMentions(params *twitter.MentionTimelineParams) ([]twitter.Tweet, error) {
	tweets, _, err := t.Client.Timelines.MentionTimeline(params)
	return tweets, err
}
