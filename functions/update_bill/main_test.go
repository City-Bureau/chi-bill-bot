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
		Data:    `{"id": "", "identifier": "R2019-11", "actions": []}`,
	}
	ocdBill := models.OCDBill{
		Actions: []models.Action{},
	}
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	UpdateBill(bill, ocdBill, snsMock)
	snsMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything, "save_bill")
}

func TestUpdateBillNilNextRunSendsTweet(t *testing.T) {
	snsMock := new(mocks.SNSClientMock)
	bill := models.Bill{
		BillID: "R201911",
		Data:   `{"actions": []}`,
	}
	ocdBill := models.OCDBill{
		Actions: []models.Action{},
	}
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	UpdateBill(bill, ocdBill, snsMock)
	snsMock.AssertExpectations(t)
	snsMock.AssertNumberOfCalls(t, "Publish", 2)
}

func TestUpdateBillNewActionsSendsTweet(t *testing.T) {
	snsMock := new(mocks.SNSClientMock)
	now := time.Now()
	bill := models.Bill{
		BillID:  "R201911",
		NextRun: &now,
		Data:    `{"actions": []}`,
	}
	ocdBill := models.OCDBill{
		Actions: []models.Action{
			models.Action{
				Date:           "2019-01-01",
				Description:    "",
				Classification: []string{},
			},
		},
	}
	snsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	UpdateBill(bill, ocdBill, snsMock)
	snsMock.AssertExpectations(t)
	snsMock.AssertNumberOfCalls(t, "Publish", 2)
}
