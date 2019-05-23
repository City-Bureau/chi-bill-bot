package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/City-Bureau/chi-bill-bot/pkg/svc"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func QueryBills(db *gorm.DB, snsClient svc.SNSType) error {
	var bills []models.Bill
	db.Limit(10).Find(
		&bills,
		"active = true AND next_run <= ? AND bill_id != '' AND bill_id IS NOT NULL",
		time.Now(),
	)

	// Iterate through mentions, publishing each to SNS topic
	for _, bill := range bills {
		billData, _ := bill.GetOCDBillData()
		ocdBill := bill.GetOCDBill()
		if len(ocdBill.Actions) <= len(billData.Actions) {
			// Set NextRun and exit because it hasn't changed
			nextRun := time.Now().Add(24 * time.Hour)
			bill.NextRun = &nextRun
			db.Save(&bill)
			// TODO: Add a check for dead bills with no activity in certain duration
			continue
		}
		billDataJson, _ := json.Marshal(billData)
		bill.Data = string(billDataJson)
		// Update with new data and publish the bill ID
		billJson, _ := json.Marshal(bill)
		snsClient.Publish(string(billJson), os.Getenv("SNS_TOPIC_ARN"))
	}
	return nil
}

func handler(request events.CloudWatchEvent) error {
	db, err := gorm.Open("mysql", "CONN")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	return QueryBills(db, svc.NewSNSClient())
}

func main() {
	lambda.Start(handler)
}
