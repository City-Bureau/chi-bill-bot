package main

import (
	"fmt"
	"os"

	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func handler(request events.CloudWatchEvent) error {
	db, _ := gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?parseTime=true",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	// db.DropTable(&models.Bill{})
	db.AutoMigrate(&models.Bill{})
	defer db.Close()

	return nil
}

func main() {
	lambda.Start(handler)
}
