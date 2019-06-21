package main

import (
	"testing"
	"time"

	"github.com/City-Bureau/chi-bill-bot/pkg/mocks"
	"github.com/City-Bureau/chi-bill-bot/pkg/models"
	"github.com/stretchr/testify/mock"
)

func TestUpdateBillIgnoresNoChanges(t *testing.T) {
	snsMock := new(mocks.SNSClientMock)
	now := time.Now()
	bill := models.Bill{
		BillID:  "R201911",
		NextRun: &now,
		Data:    `[]`,
	}
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	UpdateBill(bill, []models.LegistarAction{}, snsMock)
	snsMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything, "save_bill")
}

func TestUpdateBillNilNextRunSendsTweet(t *testing.T) {
	snsMock := new(mocks.SNSClientMock)
	bill := models.Bill{
		BillID: "R201911",
		Data:   `[]`,
	}
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	UpdateBill(bill, []models.LegistarAction{}, snsMock)
	snsMock.AssertExpectations(t)
	snsMock.AssertNumberOfCalls(t, "Publish", 2)
}

func TestUpdateBillNewActionsSendsTweet(t *testing.T) {
	snsMock := new(mocks.SNSClientMock)
	now := time.Now()
	bill := models.Bill{
		BillID:  "R201911",
		NextRun: &now,
		Data:    `[]`,
	}
	actionDate, _ := time.Parse("2006-01-02", "2019-01-01")
	actions := []models.LegistarAction{
		models.LegistarAction{Date: actionDate, Action: "introduction", Actor: "Test"},
	}
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	UpdateBill(bill, actions, snsMock)
	snsMock.AssertExpectations(t)
	snsMock.AssertNumberOfCalls(t, "Publish", 2)
}
