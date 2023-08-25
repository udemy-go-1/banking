package app

import (
	"github.com/aliciatay-zls/banking/domain"
	"github.com/aliciatay-zls/banking/logger"
	"github.com/aliciatay-zls/banking/service"
	"github.com/gorilla/mux"
	"net/http"
)

func Start() {
	router := mux.NewRouter()

	ch := CustomerHandlers{service.NewCustomerService(domain.NewCustomerRepositoryDb())}

	router.HandleFunc("/customers", ch.customersHandler).Methods(http.MethodGet)
	router.HandleFunc("/customers/{customer_id:[0-9]+}", ch.customerIdHandler).Methods(http.MethodGet)

	err := http.ListenAndServe("localhost:8080", router)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

//Notes
//create custom multiplexer/handler using mux package

//wiring = create REST handler
//a. create actual repo using the DB/adapter (initializes a repo with data queried from db)
//b. create service by passing in the repo (initialize its repo field with the actual repo)
//c. create instance of REST handler by passing in the service (initializes its service field)

//using the custom multiplexer, register route (pattern (url) --> handler method (writes response))
//gorilla mux: paths can have variables + if given vars don't match regex, mux sends error, req doesn't reach app

//start and run server
//listen on localhost and pass multiplexer to Serve()
