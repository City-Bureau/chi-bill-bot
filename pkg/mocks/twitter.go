package mocks

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/stretchr/testify/mock"
)

type TwitterMock struct {
	mock.Mock
}

func (m *TwitterMock) PostTweet(tweet string, params *twitter.StatusUpdateParams) error {
	m.Called(tweet, params)
	return nil
}

func (m *TwitterMock) GetMentions(params *twitter.MentionTimelineParams) ([]twitter.Tweet, error) {
	args := m.Called(params)
	tweets := args.Get(0).([]twitter.Tweet)
	return tweets, nil
}
