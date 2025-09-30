package main

import (
	"net/http"
	"panel-service/src/lib/logger"
)

func main() {
	logger.Init()

	logger.Log.Info("Starting server on port 8081")
	err := http.ListenAndServe(":8081", routes())
	if err != nil {
		logger.Log.Fatal(err)
	}
}
