package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"
	"github.com/getsentry/sentry-go"
	"github.com/jinzhu/gorm"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func handler(request events.CloudWatchEvent) error {
	db, err := gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?parseTime=true",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
		return err
	}
	defer db.Close()

	var bills []models.Bill
	snsClient := svc.NewSNSClient()
	db.Limit(5).Find(
		&bills,
		"active = true AND (next_run <= ? OR next_run IS NULL) AND bill_id IS NOT NULL AND bill_id != ''",
		time.Now(),
	)

	for _, bill := range bills {
		log.Println(bill.BillID)
		// Log errors but don't exit since we can just ignore them here
		billJSON, err := json.Marshal(bill)
		if err != nil {
			log.Println(err)
		}
		err = snsClient.Publish(string(billJSON), os.Getenv("SNS_TOPIC_ARN"), "update_bill")
		if err != nil {
			sentry.CaptureException(err)
			log.Println(err)
		}
	}
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
