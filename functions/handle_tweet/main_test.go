package main

import (
	"testing"

	"github.com/City-Bureau/chi-bill-bot/pkg/mocks"
	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stretchr/testify/mock"
)

func TestHandleTweetExits(t *testing.T) {
	db, dbMock, _ := sqlmock.New()
	DB, _ := gorm.Open("mysql", db)

	snsMock := new(mocks.SNSClientMock)
	var bill models.Bill
	var tweetID int64 = 1234
	bill = models.Bill{
		BillID:    "O20101",
		TweetID:   &tweetID,
		TweetText: "O20101",
	}
	dbMock.ExpectQuery("SELECT (.+) FROM (.+) WHERE (.+) LIMIT 1").
		WithArgs(bill.TweetID).
		WillReturnRows(sqlmock.NewRows([]string{"pk", "tweet_id"}).AddRow(1, 1234))
	handleTweet(&bill, DB, snsMock)
	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	snsMock.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, "save_bill")
}

func TestHandleTweetEmptyBillID(t *testing.T) {
	db, dbMock, _ := sqlmock.New()
	DB, _ := gorm.Open("mysql", db)

	snsMock := new(mocks.SNSClientMock)
	var bill models.Bill
	var tweetID int64 = 1
	bill = models.Bill{
		TweetID:   &tweetID,
		TweetText: "",
	}
	dbMock.ExpectQuery("SELECT (.+) FROM (.+) WHERE (.+) LIMIT 1").
		WithArgs(bill.TweetID).
		WillReturnError(gorm.ErrRecordNotFound)
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handleTweet(&bill, DB, snsMock)
	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	snsMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything, "save_bill")
	snsMock.AssertNumberOfCalls(t, "Publish", 1)
}
