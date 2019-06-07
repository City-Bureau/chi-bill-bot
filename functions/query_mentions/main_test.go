package main

import (
	"encoding/json"
	"testing"

	"github.com/City-Bureau/chi-bill-bot/pkg/mocks"
	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/stretchr/testify/mock"
)

func TestQueryMentionsParsesTweetId(t *testing.T) {
	var params twitter.MentionTimelineParams
	twttrMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	twttrMock.On("GetMentions", mock.Anything).Return([]twitter.Tweet{}, nil)
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	QueryMentions("", twttrMock, snsMock)
	twttrMock.AssertCalled(t, "GetMentions", &params)

	QueryMentions("1234", twttrMock, snsMock)
	twttrMock.AssertCalled(t, "GetMentions", &twitter.MentionTimelineParams{SinceID: 1234})
}

func TestQueryMentionsChecksBillID(t *testing.T) {
	tweets := []twitter.Tweet{
		twitter.Tweet{ID: 1, FullText: "Testing bill"},
	}
	twttrMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	twttrMock.On("GetMentions", mock.Anything).Return(tweets, nil)
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	tweetBill := models.Bill{
		BillID:      "",
		TweetID:     &tweets[0].ID,
		TweetText:   "Testing bill",
		LastTweetID: &tweets[0].ID,
	}
	tweetBillJson, _ := json.Marshal(tweetBill)
	QueryMentions("", twttrMock, snsMock)
	snsMock.AssertCalled(t, "Publish", string(tweetBillJson), mock.Anything, "handle_tweet")
}
