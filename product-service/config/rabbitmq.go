package config

import (
	"log"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	Id          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Price       int32              `json:"price,omitempty" bson:"price,omitempty"`
}

type Order struct {
	Product []Product
	Total   int
}

type Conn struct {
	Channel *amqp.Channel
}

func GetConn() (Conn, error) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq/")
	FailOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	return Conn{
		Channel: ch,
	}, err
}

func (conn Conn) PublishQueue(payload []byte, queueName string) {
	ch := conn.Channel
	q, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	FailOnError(err, "Failed to declare a queue")

	body := string(payload)
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		},
	)
	log.Printf(" [x] Sent %s", body)
	FailOnError(err, "Failed to publish a message")
}
