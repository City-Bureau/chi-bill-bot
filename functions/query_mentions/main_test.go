package main

import (
	"log"
	"testing"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/mocks"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/stretchr/testify/mock"
)

func TestQueryMentionsIgnoresEmptyBillID(t *testing.T) {
	tweets := []twitter.Tweet{
		twitter.Tweet{
			ID:        1,
			Text:      "Testing bill",
			User:      &twitter.User{ScreenName: "testuser"},
			CreatedAt: time.Now().Format(time.RubyDate),
		},
	}
	twttrMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	twttrMock.On("GetMentions", mock.Anything).Return(tweets, nil)
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	queryMentions(twttrMock, snsMock)
	snsMock.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, "handle_tweet")
}

func TestQueryMentionsIgnoresOldTweet(t *testing.T) {
	tweets := []twitter.Tweet{
		twitter.Tweet{
			ID:        1,
			Text:      "@chicagoledger O2010-11 Testing bill",
			User:      &twitter.User{ScreenName: "testuser"},
			CreatedAt: time.Now().UTC().Add(time.Hour * -72).Format(time.RubyDate),
		},
	}
	twttrMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	twttrMock.On("GetMentions", mock.Anything).Return(tweets, nil)
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	queryMentions(twttrMock, snsMock)
	snsMock.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, "handle_tweet")
}

func TestQueryMentionsTweetsBill(t *testing.T) {
	log.Printf(time.Now().UTC().String())
	tweets := []twitter.Tweet{
		twitter.Tweet{
			ID:        1,
			Text:      "@chicagoledger O2010-11 Testing bill",
			User:      &twitter.User{ScreenName: "testuser"},
			CreatedAt: time.Now().Format(time.RubyDate),
		},
	}
	twttrMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	twttrMock.On("GetMentions", mock.Anything).Return(tweets, nil)
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	queryMentions(twttrMock, snsMock)
	snsMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything, "handle_tweet")
}
