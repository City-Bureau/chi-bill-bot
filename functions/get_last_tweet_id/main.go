package main

import (
	"fmt"
	"log"
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

	db.Order("last_tweet_id desc").First(&bill)
	snsClient := svc.NewSNSClient()
	return snsClient.Publish(string(*bill.LastTweetID), os.Getenv("SNS_TOPIC_ARN"), "query_mentions")
}

func main() {
	lambda.Start(handler)
}
