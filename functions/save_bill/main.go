package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/getsentry/sentry-go"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func handler(request events.SNSEvent) error {
	if len(request.Records) < 0 {
		return nil
	}
	message := request.Records[0].SNS.Message
	db, err := gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?parseTime=true",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer db.Close()

	var bill models.Bill
	err = json.Unmarshal([]byte(message), &bill)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
		return err
	}

	_ = db.Save(&bill)
	return nil
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
