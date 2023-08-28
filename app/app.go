package app

import (
	"fmt"
	"github.com/aliciatay-zls/banking/domain"
	"github.com/aliciatay-zls/banking/logger"
	"github.com/aliciatay-zls/banking/service"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"net/http"
	"os"
	"time"
)

func checkEnvVars() {
	envVars := []string{
		"SERVER_ADDRESS",
		"SERVER_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_ADDRESS",
		"DB_PORT",
		"DB_NAME",
	}

	for _, key := range envVars {
		if os.Getenv(key) == "" {
			logger.Fatal(fmt.Sprintf("Environment variable %s was not defined", key))
		}
	}
}

func Start() {
	checkEnvVars()

	router := mux.NewRouter()

	dbClient := getDbClient()
	customerRepositoryDb := domain.NewCustomerRepositoryDb(dbClient)
	accountRepositoryDb := domain.NewAccountRepositoryDb(dbClient)
	ch := CustomerHandlers{service.NewCustomerService(customerRepositoryDb)}
	ah := AccountHandler{service.NewAccountService(accountRepositoryDb)}

	router.HandleFunc("/customers", ch.customersHandler).Methods(http.MethodGet)
	router.HandleFunc("/customers/{customer_id:[0-9]+}", ch.customerIdHandler).Methods(http.MethodGet)
	router.HandleFunc("/customers/{customer_id:[0-9]+}/account", ah.newAccountHandler).Methods(http.MethodPost)

	address := os.Getenv("SERVER_ADDRESS")
	port := os.Getenv("SERVER_PORT")
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), router)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func getDbClient() *sqlx.DB {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbAddress := os.Getenv("DB_ADDRESS")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbAddress, dbPort, dbName)
	db, err := sqlx.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}

//Notes
//once the app is started, check that environment variables required for the app to function have been set

//create custom multiplexer/handler using mux package

//wiring = create REST handler
//0. connect to the database/get a database handle
//1. create instance of DB/adapter (initialize adapter with database handle)
//2. create instance of service by passing in the adapter as the repo implementation
//(initialize service's repo field with adapter)
//3. create instance of REST handler by passing in the service (initialize handler's service field)

//using the custom multiplexer, register route (pattern (url) --> handler method (writes response))
//gorilla mux: paths can have variables + if given vars don't match regex, mux sends error, req doesn't reach app

//start and run server
//listen on localhost and pass multiplexer to Serve()
