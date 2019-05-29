package mocks

import (
	"github.com/stretchr/testify/mock"
)

type SNSClientMock struct {
	mock.Mock
}

func (m *SNSClientMock) Publish(message string, topicArn string, feed string) error {
	m.Called(message, topicArn, feed)
	return nil
}
