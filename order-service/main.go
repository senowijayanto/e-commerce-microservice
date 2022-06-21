package main

import (
	"log"
	"order-service/config"
)

func main() {
	log.Println("Order-Service is running!")

	conn, err := config.GetConn()
	config.FailOnError(err, "Failed connected RabbitMQ")

	conn.ConsumeQueue("ORDER")
}
