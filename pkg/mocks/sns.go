package mocks

import (
	"github.com/stretchr/testify/mock"
)

type SNSClientMock struct {
	mock.Mock
}

func (m *SNSClientMock) Publish(message string, topicArn string) error {
	m.Called(message, topicArn)
	return nil
}
