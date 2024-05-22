package main

import (
	"github.com/udemy-go-1/banking-lib/formValidator"
	"github.com/udemy-go-1/banking-lib/logger"
	"github.com/udemy-go-1/banking/backend/app"
)

func main() {
	logger.Info("Starting the app...")
	formValidator.Create()
	app.Start()
}
