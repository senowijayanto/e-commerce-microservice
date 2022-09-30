package main

import (
	"encoding/json"
	"log"
	"net/http"
	"order-service/config"

	"github.com/gorilla/mux"
)

type Response struct {
	Status  bool   `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func HealthCheck(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	result := Response{
		Status:  true,
		Code:    http.StatusOK,
		Message: "All checked!",
	}

	json.NewEncoder(response).Encode(result)
}

func main() {
	log.Println("Order-Service is running!")

	conn, err := config.GetConn()
	config.FailOnError(err, "Failed connected RabbitMQ")

	conn.ConsumeQueue("ORDER")

	router := mux.NewRouter()

	router.HandleFunc("/order", HealthCheck).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
