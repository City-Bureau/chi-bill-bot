package svc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNSType interface {
	Publish(string, string) error
}

type SNSClient struct {
	Client *sns.SNS
}

func NewSNSClient() *SNSClient {
	client := sns.New(session.New())
	return &SNSClient{Client: client}
}

func (c *SNSClient) Publish(message string, topicArn string) error {
	_, err := c.Client.Publish(&sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(topicArn),
	})
	return err
}
