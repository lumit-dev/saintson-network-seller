package routes

import (
	"net/http"
	logger "paymentserv/src/lib/logger"
)

func Listen() {
	logger.Log.Info("starting server on port 8082")
	err := http.ListenAndServe(":8082", routes())
	if err != nil {
		logger.Log.Fatal(err)
	}

}
