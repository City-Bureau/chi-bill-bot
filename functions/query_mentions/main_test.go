package main

import (
	"log"
	"testing"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/mocks"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/stretchr/testify/mock"
)

const TIME_FORMAT = "Mon Jan 2 15:04:05 -0700 2006"

func TestQueryMentionsIgnoresEmptyBillID(t *testing.T) {
	tweets := []twitter.Tweet{
		twitter.Tweet{
			ID:        1,
			Text:      "Testing bill",
			CreatedAt: time.Now().Format(TIME_FORMAT),
		},
	}
	twttrMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	twttrMock.On("GetMentions", mock.Anything).Return(tweets, nil)
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	QueryMentions(twttrMock, snsMock)
	snsMock.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, "handle_tweet")
}

func TestQueryMentionsIgnoresOldTweet(t *testing.T) {
	tweets := []twitter.Tweet{
		twitter.Tweet{
			ID:        1,
			Text:      "@chicagoledger O2010-11 Testing bill",
			CreatedAt: time.Now().Add(time.Hour * -72).Format(TIME_FORMAT),
		},
	}
	twttrMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	twttrMock.On("GetMentions", mock.Anything).Return(tweets, nil)
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	QueryMentions(twttrMock, snsMock)
	snsMock.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, "handle_tweet")
}

func TestQueryMentionsTweetsBill(t *testing.T) {
	log.Printf(time.Now().UTC().String())
	tweets := []twitter.Tweet{
		twitter.Tweet{
			ID:        1,
			Text:      "@chicagoledger O2010-11 Testing bill",
			CreatedAt: time.Now().Format(TIME_FORMAT),
		},
	}
	twttrMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	twttrMock.On("GetMentions", mock.Anything).Return(tweets, nil)
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	QueryMentions(twttrMock, snsMock)
	snsMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything, "handle_tweet")
}
