package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
	Id          interface{} `json:"id" bson:"id"`
	Name        string      `json:"name" bson:"name"`
	Description string      `json:"description" bson:"description"`
	Price       int32       `json:"price" bson:"price"`
}

type ProductResponse struct {
	Id          interface{} `json:"id" bson:"id"`
	Name        string      `json:"name" bson:"name"`
	Description string      `json:"description" bson:"description"`
	Price       int32       `json:"price" bson:"price"`
}

var (
	SECRET_KEY                 = []byte("secretkey")
	MONGO_URI                  = "mongodb://admin:secret@localhost:27019"
	MONGO_DATABASE             = "productDB"
	MONGO_POOL_MIN             = 10
	MONGO_POOL_MAX             = 100
	MONGO_MAX_IDLE_TIME_SECOND = 60
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func NewMongoContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func NewMongoDatabase() *mongo.Database {
	ctx, cancel := NewMongoContext()
	defer cancel()

	option := options.Client().
		ApplyURI(MONGO_URI).
		SetMinPoolSize(uint64(MONGO_POOL_MIN)).
		SetMaxPoolSize(uint64(MONGO_POOL_MAX)).
		SetMaxConnIdleTime(time.Duration(MONGO_MAX_IDLE_TIME_SECOND) * time.Second)

	client, err := mongo.NewClient(option)
	if err != nil {
		log.Panic(err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Panic(err)
	}

	database := client.Database(MONGO_DATABASE)
	return database
}

func Create(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := NewMongoContext()
	defer cancel()

	response.Header().Set("Content-Type", "application/json")

	var product Product
	json.NewDecoder(request.Body).Decode(&product)

	collection := NewMongoDatabase().Collection("product")

	// Save product
	res, err := collection.InsertOne(ctx, product)
	if err != nil {
		log.Panic(err)
	}

	result := ProductResponse{
		Id:          res.InsertedID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}

	json.NewEncoder(response).Encode(result)
}

func Buy(response http.ResponseWriter, request *http.Request) {
	// ctx, cancel := NewMongoContext()
	// defer cancel()

	// Get product from body request
	var product Product
	// var dbProduct Product
	json.NewDecoder(request.Body).Decode(&product)

	// Setup connection RabbitMQ
	// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	// failOnError(err, "Failed to connect to RabbitMQ")

	fmt.Println(reflect.ValueOf(product.Id).Len())
	// Find product by id
	// collection := NewMongoDatabase().Collection("product")
	// // err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": product.Id}}).All(&dbProduct)
	// cursor, err := collection.Find(context.TODO(), bson.M{"_id": bson.M{"$in": product.Id}})
	// failOnError(err, "Product Not Found!")

	// err = cursor.All(context.TODO(), &dbProduct)
	// failOnError(err, "Failed!")

	// fmt.Println(dbProduct)
	// // Create Channel
	// ch, err := conn.Channel()
	// failOnError(err, "Failed to connect to RabbitMQ")
	// defer ch.Close()

	// q, err := ch.QueueDeclare(
	// 	"ORDER", // name
	// 	false,   // durable
	// 	false,   // delete when unused
	// 	false,   // exclusive
	// 	false,   // no-wait
	// 	nil,     // arguments
	// )
	// log.Println(q)

	// failOnError(err, "Failed to connect to RabbitMQ")

	// dataProduct := &Product{
	// 	Id:          dbProduct.Id,
	// 	Name:        dbProduct.Name,
	// 	Description: dbProduct.Description,
	// 	Price:       dbProduct.Price,
	// }

	// body, err := json.Marshal(dataProduct)
	// failOnError(err, "Failed convert product!")

	// err = ch.Publish(
	// 	"",
	// 	"ORDER",
	// 	false,
	// 	false,
	// 	amqp.Publishing{
	// 		ContentType: "application/json",
	// 		Body:        []byte(body),
	// 	},
	// )

	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println("Successfully Published Message to Queue")
}

func main() {
	log.Println("Product-Service is running!")

	router := mux.NewRouter()

	router.HandleFunc("/product/create", Create).Methods("POST")
	router.HandleFunc("/product/buy", Buy).Methods("POST")

	log.Fatal(http.ListenAndServe("localhost:8082", router))
}
