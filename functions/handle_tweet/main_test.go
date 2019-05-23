package main

import (
	"database/sql"
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

	twitterMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	var bill models.Bill
	bill = models.Bill{
		BillID:    "O20101",
		TweetID:   sql.NullInt64{Int64: 1234},
		TweetText: "O20101",
	}
	dbMock.ExpectQuery("SELECT (.+) FROM (.+) WHERE (.+) LIMIT 1").
		WithArgs(bill.TweetID).
		WillReturnRows(sqlmock.NewRows([]string{"pk", "tweet_id"}).AddRow(1, 1234))

	HandleTweet(&bill, DB, twitterMock, snsMock)
	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	twitterMock.AssertNotCalled(t, "PostTweet", mock.Anything, mock.Anything)
}

func TestHandleTweetEmptyBillID(t *testing.T) {
	db, dbMock, _ := sqlmock.New()
	DB, _ := gorm.Open("mysql", db)

	twitterMock := new(mocks.TwitterMock)
	snsMock := new(mocks.SNSClientMock)
	var bill models.Bill
	bill = models.Bill{
		TweetID:   sql.NullInt64{Int64: 1},
		TweetText: "",
	}
	dbMock.ExpectQuery("SELECT (.+) FROM (.+) WHERE (.+) LIMIT 1").
		WithArgs(bill.TweetID).
		WillReturnError(gorm.ErrRecordNotFound)
	twitterMock.On("PostTweet", mock.Anything, mock.Anything)
	HandleTweet(&bill, DB, twitterMock, snsMock)
	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	twitterMock.AssertExpectations(t)
}
