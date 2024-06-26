package app

import (
	"bytes"
	"github.com/aliciatay-zls/banking-lib/errs"
	"github.com/aliciatay-zls/banking-lib/formValidator"
	"github.com/aliciatay-zls/banking-lib/logger"
	"github.com/aliciatay-zls/banking/backend/dto"
	"github.com/aliciatay-zls/banking/backend/mocks/service"
	"github.com/gorilla/mux"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test common variables and inputs
var mockAccountService *service.MockAccountService
var ah AccountHandler

var dummyAmount float64 = 6000

var dummyAccountType = dto.AccountTypeSaving
var dummyTransactionType = dto.TransactionTypeDeposit

const dummyDate = "2006-01-02 15:04:05"

const getAccountsPath = "/customers/{customer_id:[0-9]+}"
const dummyGetAccountsPath = "/customers/2"

const newAccountPath = "/customers/{customer_id:[0-9]+}/account/new"
const dummyNewAccountPath = "/customers/2/account/new"
const dummyNewAccountRequestPayload = `{"account_type": "saving", "amount": 6000}`
const dummyAccountId = "1977"

const newTransactionPath = "/customers/{customer_id:[0-9]+}/account/{account_id:[0-9]+}"
const dummyNewTransactionPath = "/customers/2/account/1977"
const dummyNewTransactionPayload = `{"transaction_type": "deposit", "amount": 6000}`
const dummyTransactionId = "7791"
const dummyBalance float64 = 12000

func init() {
	formValidator.Create()
}

func setupAccountHandlerTest(t *testing.T, path string, payload string) func() {
	ctrl := gomock.NewController(t)
	mockAccountService = service.NewMockAccountService(ctrl)
	ah = AccountHandler{mockAccountService}

	router = mux.NewRouter()

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer([]byte(payload)))

	return func() {
		router = nil
		recorder = nil
		request = nil
		defer ctrl.Finish()
	}
}

// getDefaultDummyNewAccountRequestObject returns a dto.NewAccountRequest for the customer with id 2 wanting to open
// a saving account of amount 6000
func getDefaultDummyNewAccountRequestObject() dto.NewAccountRequest {
	return dto.NewAccountRequest{
		CustomerId:  dummyCustomerId,
		AccountType: dummyAccountType,
		Amount:      dummyAmount,
	}
}

// getDefaultDummyNewTransactionRequestObject returns a dto.TransactionRequest for the customer with id 2 wanting to
// make a deposit of amount 6000 on the account with id 1977
func getDefaultDummyNewTransactionRequestObject() dto.TransactionRequest {
	return dto.TransactionRequest{
		AccountId:       dummyAccountId,
		Amount:          dummyAmount,
		TransactionType: dummyTransactionType,
		CustomerId:      dummyCustomerId,
	}
}

func TestAccountHandler_accountsHandler_respondsWith_errorStatusCode_when_service_fails(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyGetAccountsPath, "")
	defer teardown()
	request = httptest.NewRequest(http.MethodGet, dummyGetAccountsPath, nil) //override
	router.HandleFunc(getAccountsPath, ah.accountsHandler).Methods(http.MethodGet)

	dummyAppErr := errs.NewUnexpectedError("some error message")
	mockAccountService.EXPECT().GetAllAccounts(dummyCustomerId).Return(nil, dummyAppErr)

	//Act
	router.ServeHTTP(recorder, request)

	//Assert
	if recorder.Result().StatusCode != dummyAppErr.Code {
		t.Errorf("Expected status code %d but got %d", dummyAppErr.Code, recorder.Result().StatusCode)
	}
	actualResponse, _ := io.ReadAll(recorder.Result().Body)
	if !strings.Contains(string(actualResponse), dummyAppErr.Message) {
		t.Errorf("Expected response to contain %s but got: %s", dummyAppErr.Message, actualResponse)
	}
}

func TestAccountHandler_accountsHandler_respondsWith_accountsAndStatusCode200_when_service_succeeds(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyGetAccountsPath, "")
	defer teardown()
	request = httptest.NewRequest(http.MethodGet, dummyGetAccountsPath, nil) //override
	router.HandleFunc(getAccountsPath, ah.accountsHandler).Methods(http.MethodGet)

	dummyAccounts := []dto.AccountResponse{
		{dummyAccountId, dummyDate, dummyAccountType, dummyAmount},
		{"1980", dummyDate, dto.AccountTypeChecking, 7000},
	}
	mockAccountService.EXPECT().GetAllAccounts(dummyCustomerId).Return(dummyAccounts, nil)

	expectedStatusCode := http.StatusOK

	//Act
	router.ServeHTTP(recorder, request)

	//Assert
	if recorder.Result().StatusCode != expectedStatusCode {
		t.Errorf("Expected status code %d but got %d", expectedStatusCode, recorder.Result().StatusCode)
	}
	actualResponse, _ := io.ReadAll(recorder.Result().Body)
	for k, _ := range dummyAccounts {
		if !strings.Contains(string(actualResponse), dummyAccounts[k].AccountId) {
			t.Errorf("Expected response to contain account with id %s but did not", dummyAccounts[k].AccountId)
		}
	}
}

func TestAccountHandler_newAccountHandler_respondsWith_errorStatusCode_when_payload_malformed(t *testing.T) {
	//Arrange
	badPayload := `{"customer_id": "2"", account_type": "saving", "amount": "string instead of number"}`
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

func TestAccountHandler_newAccountHandler_respondsWith_newAccountAndStatusCode200_when_service_succeeds(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyNewAccountPath, dummyNewAccountRequestPayload)
	defer teardown()
	router.HandleFunc(newAccountPath, ah.newAccountHandler).Methods(http.MethodPost)

	dummyNewAccountRequestObject := getDefaultDummyNewAccountRequestObject()
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

func TestAccountHandler_newAccountHandler_respondsWith_errorStatusCode_when_service_fails(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyNewAccountPath, dummyNewAccountRequestPayload)
	defer teardown()
	router.HandleFunc(newAccountPath, ah.newAccountHandler).Methods(http.MethodPost)

	dummyNewAccountRequestObject := getDefaultDummyNewAccountRequestObject()
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

func TestAccountHandler_transactionHandler_respondsWith_errorStatusCode_when_payload_malformed(t *testing.T) {
	//Arrange
	badPayload := `{"account_id": "1977", "customer_id": "2", "transaction_type": "deposit", "amount": "string instead of number"}`
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

func TestAccountHandler_transactionHandler_respondsWith_newTransactionAndStatusCode200_when_service_succeeds(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyNewTransactionPath, dummyNewTransactionPayload)
	defer teardown()
	router.HandleFunc(newTransactionPath, ah.transactionHandler)

	dummyNewTransactionRequestObject := getDefaultDummyNewTransactionRequestObject()
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

func TestAccountHandler_transactionHandler_respondsWith_errorStatusCode_when_service_fails(t *testing.T) {
	//Arrange
	teardown := setupAccountHandlerTest(t, dummyNewTransactionPath, dummyNewTransactionPayload)
	defer teardown()
	router.HandleFunc(newTransactionPath, ah.transactionHandler)

	dummyNewTransactionRequestObject := getDefaultDummyNewTransactionRequestObject()
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
