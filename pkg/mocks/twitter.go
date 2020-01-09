package mocks

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/stretchr/testify/mock"
)

// TwitterMock is a struct for mocking the Twitter service
type TwitterMock struct {
	mock.Mock
}

// PostTweet mocks posting a tweet on Twitter
func (m *TwitterMock) PostTweet(tweet string, params *twitter.StatusUpdateParams) error {
	m.Called(tweet, params)
	return nil
}

// GetMentions mocks GetMentions on the Twitter service
func (m *TwitterMock) GetMentions(params *twitter.MentionTimelineParams) ([]twitter.Tweet, error) {
	args := m.Called(params)
	tweets := args.Get(0).([]twitter.Tweet)
	return tweets, nil
}
