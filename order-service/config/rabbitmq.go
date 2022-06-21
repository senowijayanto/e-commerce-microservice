package config

import (
	"encoding/json"
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
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	FailOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	return Conn{
		Channel: ch,
	}, err
}

func (conn Conn) ConsumeQueue(queueName string) {
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

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	FailOnError(err, "Failed to register a consumer")

	forever := make(chan bool)
	var products []Product

	go func() {
		for d := range msgs {
			err := json.Unmarshal(d.Body, &products)
			FailOnError(err, "Error Unmarshal")
			createOrder(products)
		}
	}()

	log.Printf(" [*] Waiting for messages...")
	<-forever
}

func createOrder(products []Product) {
	ctx, cancel := NewMongoContext()
	defer cancel()

	total := 0
	for i := 0; i < len(products); i++ {
		total += int(products[i].Price)
	}
	order := Order{
		Product: products,
		Total:   total,
	}

	collection := NewMongoDatabase().Collection("order")

	// Save order
	_, err := collection.InsertOne(ctx, order)
	FailOnError(err, "Failed insert data!")

	log.Println("Insert order success!")
}
