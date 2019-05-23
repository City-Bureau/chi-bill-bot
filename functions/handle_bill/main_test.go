package main

import (
	"testing"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/mocks"
	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stretchr/testify/mock"
)

func TestHandleBill(t *testing.T) {
	db, _, _ := sqlmock.New()
	DB, _ := gorm.Open("mysql", db)

	twitterMock := new(mocks.TwitterMock)
	twitterMock.On("PostTweet", "Test tweet", mock.Anything).Return(nil)
	snsMock := new(mocks.SNSClientMock)
	var bill models.Bill
	HandleBill("Test tweet", &bill, DB, twitterMock, snsMock)

	// Should post tweet and set NextRun
	if bill.NextRun.Before(time.Now()) {
		t.Errorf("NextRun should be set in the future")
	}
	twitterMock.AssertExpectations(t)
}
