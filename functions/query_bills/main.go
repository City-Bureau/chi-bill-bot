package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"
	"github.com/jinzhu/gorm"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

	var bills []models.Bill
	db.Limit(5).Find(
		&bills,
		"active = true AND next_run <= ? AND bill_id IS NOT NULL",
		time.Now(),
	)

	billsJson, _ := json.Marshal(bills)
	snsClient := svc.NewSNSClient()
	snsClient.Publish(string(billsJson), os.Getenv("SNS_TOPIC_ARN"), "update_bills")
	return nil
}

func main() {
	lambda.Start(handler)
}
