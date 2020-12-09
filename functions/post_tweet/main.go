package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/svc"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Event should include bill and params
func handler(request events.SNSEvent) error {
	if len(request.Records) < 0 {
		return nil
	}
	message := request.Records[0].SNS.Message

	var data svc.TweetData
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	twttr := svc.NewTwitterClient()
	err = twttr.PostTweet(data.Text, &data.Params)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			log.Println("Status is a duplicate, ignoring error")
			return nil
		}
	}
	return err
}

func main() {
	_ = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
		Transport: &sentry.HTTPSyncTransport{
			Timeout: 5 * time.Second,
		},
	})

	lambda.Start(handler)
}
