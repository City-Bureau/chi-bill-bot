package main

import (
	"fmt"
	"os"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func handler(request events.CloudWatchEvent) error {
	db, err := gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var bill models.Bill

	db.Order("last_tweet_id desc").First(&bill)
	snsClient := svc.NewSNSClient()
	snsClient.Publish(string(*bill.LastTweetID), os.Getenv("SNS_TOPIC_ARN"), "query_mentions")

	return nil
}

func main() {
	lambda.Start(handler)
}
