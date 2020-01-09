package mocks

import (
	"github.com/stretchr/testify/mock"
)

// SNSClientMock is a mock for SNS
type SNSClientMock struct {
	mock.Mock
}

// Publish mocks publishing on SNS
func (m *SNSClientMock) Publish(message string, topicArn string, feed string) error {
	m.Called(message, topicArn, feed)
	return nil
}
