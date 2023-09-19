package app

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/udemy-go-1/banking-lib/errs"
	"github.com/udemy-go-1/banking-lib/logger"
	"github.com/udemy-go-1/banking/dto"
	"github.com/udemy-go-1/banking/mocks/service"
	"go.uber.org/mock/gomock"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test variables and common inputs
// customer with id as 2 wants to open new saving account with amount 6000
// customer with id as 2 and account number as 1977 wants to make a deposit of amount 6000
var mockAccountService *service.MockAccountService
var ah AccountHandler
var dummyNewAccountRequestObject dto.NewAccountRequest
var dummyNewTransactionRequestObject dto.TransactionRequest
var dummyAccountType = dto.AccountTypeSaving
var dummyTransactionType = dto.TransactionTypeDeposit
var dummyAmount float64 = 6000

const newAccountPath = "/customers/{customer_id:[0-9]+}/account"
const newTransactionPath = "/customers/{customer_id:[0-9]+}/account/{account_id:[0-9]+}"
const dummyNewAccountPath = "/customers/2/account"
const dummyNewTransactionPath = "/customers/2/account/1977"
const dummyAccountId = "1977"
const dummyNewAccountRequestPayload = `{"account_type": "saving", "amount": 6000}`
const dummyNewTransactionPayload = `{"transaction_type": "deposit", "amount": 6000}`
const dummyTransactionId = "7791"
const dummyBalance float64 = 12000

func setupAccountHandlerTest(t *testing.T, path string, payload string) func() {
	ctrl := gomock.NewController(t)
	mockAccountService = service.NewMockAccountService(ctrl)
	ah = AccountHandler{mockAccountService}

	router = mux.NewRouter()

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer([]byte(payload)))

	dummyNewAccountRequestObject = dto.NewAccountRequest{
		CustomerId:  dummyCustomerId,
		AccountType: &dummyAccountType,
		Amount:      &dummyAmount,
	}
	dummyNewTransactionRequestObject = dto.TransactionRequest{
		AccountId:       dummyAccountId,
		Amount:          &dummyAmount,
		TransactionType: &dummyTransactionType,
		CustomerId:      dummyCustomerId,
	}

	return func() {
		router = nil
		recorder = nil
		request = nil
		defer ctrl.Finish()
	}
}

func TestAccountHandler_newAccountHandler_NoAccountWithErrorStatusCodeWhenPayloadMalformed(t *testing.T) {
	//Arrange
	badPayload := `{"account_type": "saving", "amount": "string instead of number"}`
	teardown := setupAccountHandlerTest(t, dummyNewAccountPath, badPayload)
	defer teardown()
	router.HandleFunc(newAccountPath, ah.newAccountHandler).Methods(http.MethodPost)

	expectedStatusCode := http.StatusBadRequest

	logs := logger.ReplaceWithTestLogger()
	expectedLogMessagePrefix := "Error while decoding json body of new account request: "

	//Act
	router.ServeHTTP(recorder, request)

	//Assert
	if recorder.Result().StatusCode != expectedStatusCode {
		t.Errorf("Expecting status code %d but got %d", expectedStatusCode, recorder.Result().StatusCode)
	}
	if logs.Len() != 1 {
		t.Fatalf("Expected 1 message to be logged but got %d logs", logs.Len())
	}
	actualLogMessage := logs.All()[0].Message
	if !strings.Contains(actualLogMessage, expectedLogMessagePrefix) {
		t.Errorf("Expected log message to contain \"%s\" but got log message: \"%s\"", expectedLogMessagePrefix, actualLogMessage)
	}
}

func TestAccountHandler_newAccountHandler_NoAccountWithErrorWhenPayloadFieldMissingOrNull(t *testing.T) {
	badPayloads := []string{`{"account_type": "saving"}`, `{"account_type": "saving", "amount": null}`}
	expectedStatusCode := http.StatusBadRequest
	expectedResponse := "\"Field(s) missing or null in request body: account_type, amount\""
	expectedLogMessage := "Field(s) missing or null in request body"

	for index, payload := range badPayloads {
		log.Printf("Testing with payload number %d: %s", index+1, payload)

		//Arrange
		teardown := setupAccountHandlerTest(t, dummyNewAccountPath, payload)
		router.HandleFunc(newAccountPath, ah.newAccountHandler).Methods(http.MethodPost)

		logs := logger.ReplaceWithTestLogger()

		//Act
		router.ServeHTTP(recorder, request)

		//Assert
		if recorder.Result().StatusCode != expectedStatusCode {
			t.Errorf("Expecting status code %d but got %d", expectedStatusCode, recorder.Result().StatusCode)
		}
		actualResponse, _ := io.ReadAll(recorder.Result().Body)
		if !strings.Contains(string(actualResponse), expectedResponse) {
			t.Errorf("Expecting response to contain %s but got %s", expectedResponse, actualResponse)
		}
		if logs.Len() != 1 {
			t.Fatalf("Expected 1 message to be logged but got %d logs", logs.Len())
		}
		actualLogMessage := logs.All()[0].Message
		if actualLogMessage != expectedLogMessage {
			t.Errorf("Expected log message to be \"%s\" but got \"%s\"", expectedLogMessage, actualLogMessage)
		}

		//Cleanup
		teardown()
	}

}

func TestAccountHandler_newAccountHandler_NewAccountWithStatusCode200WhenServiceSucceeds(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyNewAccountPath, dummyNewAccountRequestPayload)
	defer teardown()
	router.HandleFunc(newAccountPath, ah.newAccountHandler).Methods(http.MethodPost)

	dummyAccount := dto.NewAccountResponse{AccountId: dummyAccountId}
	mockAccountService.EXPECT().CreateNewAccount(dummyNewAccountRequestObject).Return(&dummyAccount, nil)
	expectedStatusCode := http.StatusCreated

	//Act
	router.ServeHTTP(recorder, request)

	//Assert
	if recorder.Result().StatusCode != expectedStatusCode {
		t.Errorf("Expecting status code %d but got %d", expectedStatusCode, recorder.Result().StatusCode)
	}
	actualResponse, _ := io.ReadAll(recorder.Result().Body)
	if !strings.Contains(string(actualResponse), dummyAccount.AccountId) {
		t.Errorf("Expecting response to contain %s but got %s", dummyAccount.AccountId, actualResponse)
	}
}

func TestAccountHandler_newAccountHandler_NoAccountWithErrorStatusCodeWhenServiceFails(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyNewAccountPath, dummyNewAccountRequestPayload)
	defer teardown()
	router.HandleFunc(newAccountPath, ah.newAccountHandler).Methods(http.MethodPost)

	dummyAppError := errs.NewUnexpectedError("some error message")
	mockAccountService.EXPECT().CreateNewAccount(dummyNewAccountRequestObject).Return(nil, dummyAppError)

	//Act
	router.ServeHTTP(recorder, request)

	//Assert
	if recorder.Result().StatusCode != dummyAppError.Code {
		t.Errorf("Expecting status code %d but got %d", dummyAppError.Code, recorder.Result().StatusCode)
	}
	actualResponse, _ := io.ReadAll(recorder.Result().Body)
	if !strings.Contains(string(actualResponse), dummyAppError.Message) {
		t.Errorf("Expecting response to contain %s but got %s", dummyAppError.Message, actualResponse)
	}
}

func TestAccountHandler_transactionHandler_NoTransactionWithErrorStatusCodeWhenPayloadMalformed(t *testing.T) {
	//Arrange
	badPayload := `{"transaction_type": "deposit", "amount": "string instead of number"}`
	teardown := setupAccountHandlerTest(t, dummyNewTransactionPath, badPayload)
	defer teardown()
	router.HandleFunc(newTransactionPath, ah.transactionHandler).Methods(http.MethodPost)

	expectedStatusCode := http.StatusBadRequest

	logs := logger.ReplaceWithTestLogger()
	expectedLogMessagePrefix := "Error while decoding json body of transaction request: "

	//Act
	router.ServeHTTP(recorder, request)

	//Assert
	if recorder.Result().StatusCode != expectedStatusCode {
		t.Errorf("Expecting status code %d but got %d", expectedStatusCode, recorder.Result().StatusCode)
	}
	if logs.Len() != 1 {
		t.Fatalf("Expected 1 message to be logged but got %d logs", logs.Len())
	}
	actualLogMessage := logs.All()[0].Message
	if !strings.Contains(actualLogMessage, expectedLogMessagePrefix) {
		t.Errorf("Expected log message to contain \"%s\" but got log message: \"%s\"", expectedLogMessagePrefix, actualLogMessage)
	}
}

func TestAccountHandler_transactionHandler_NoTransactionWithErrorWhenPayloadFieldMissingOrNull(t *testing.T) {
	badPayloads := []string{`{"amount": 6000}`, `{"transaction_type": null, "amount": 6000}`}
	expectedStatusCode := http.StatusBadRequest
	expectedResponse := "\"Field(s) missing or null in request body: transaction_type, amount\""
	expectedLogMessage := "Field(s) missing or null in request body"

	for index, payload := range badPayloads {
		log.Printf("Testing with payload number %d: %s", index+1, payload)

		//Arrange
		teardown := setupAccountHandlerTest(t, dummyNewTransactionPath, payload)
		router.HandleFunc(newTransactionPath, ah.transactionHandler).Methods(http.MethodPost)

		logs := logger.ReplaceWithTestLogger()

		//Act
		router.ServeHTTP(recorder, request)

		//Assert
		if recorder.Result().StatusCode != expectedStatusCode {
			t.Errorf("Expecting status code %d but got %d", expectedStatusCode, recorder.Result().StatusCode)
		}
		actualResponse, _ := io.ReadAll(recorder.Result().Body)
		if !strings.Contains(string(actualResponse), expectedResponse) {
			t.Errorf("Expecting response to contain %s but got %s", expectedResponse, actualResponse)
		}
		if logs.Len() != 1 {
			t.Fatalf("Expected 1 message to be logged but got %d logs", logs.Len())
		}
		actualLogMessage := logs.All()[0].Message
		if actualLogMessage != expectedLogMessage {
			t.Errorf("Expected log message to be \"%s\" but got \"%s\"", expectedLogMessage, actualLogMessage)
		}

		//Cleanup
		teardown()
	}

}

func TestAccountHandler_transactionHandler_NewTransactionWithStatusCode200WhenServiceSucceeds(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyNewTransactionPath, dummyNewTransactionPayload)
	defer teardown()
	router.HandleFunc(newTransactionPath, ah.transactionHandler)

	dummyTransaction := dto.TransactionResponse{TransactionId: dummyTransactionId, Balance: dummyBalance}
	mockAccountService.EXPECT().MakeTransaction(dummyNewTransactionRequestObject).Return(&dummyTransaction, nil)
	expectedStatusCode := http.StatusCreated

	//Act
	router.ServeHTTP(recorder, request)

	//Assert
	if recorder.Result().StatusCode != expectedStatusCode {
		t.Errorf("Expected status code %d but got %d", expectedStatusCode, recorder.Result().StatusCode)
	}
	actualResponse, _ := io.ReadAll(recorder.Result().Body)
	if !strings.Contains(string(actualResponse), dummyTransaction.TransactionId) {
		t.Errorf("Expecting response to contain %s but got %s", dummyTransaction.TransactionId, actualResponse)
	}
}

func TestAccountHandler_transactionHandler_NoTransactionWithErrorStatusCodeWhenServiceFails(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyNewTransactionPath, dummyNewTransactionPayload)
	defer teardown()
	router.HandleFunc(newTransactionPath, ah.transactionHandler)

	dummyAppError := errs.NewUnexpectedError("some error message")
	mockAccountService.EXPECT().MakeTransaction(dummyNewTransactionRequestObject).Return(nil, dummyAppError)

	//Act
	router.ServeHTTP(recorder, request)

	//Assert
	if recorder.Result().StatusCode != dummyAppError.Code {
		t.Errorf("Expected status code %d but got %d", dummyAppError.Code, recorder.Result().StatusCode)
	}
	actualResponse, _ := io.ReadAll(recorder.Result().Body)
	if !strings.Contains(string(actualResponse), dummyAppError.Message) {
		t.Errorf("Expecting response to contain %s but got %s", dummyAppError.Message, actualResponse)
	}
}
