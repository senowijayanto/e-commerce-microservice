package main

import (
	"encoding/json"
	"log"
	"net/http"

	"product-service/config"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	Id          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Price       int32              `json:"price,omitempty" bson:"price,omitempty"`
}

type ProductIDs struct {
	Ids []primitive.ObjectID `json:"ids"`
}

func Create(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := config.NewMongoContext()
	defer cancel()

	response.Header().Set("Content-Type", "application/json")

	var product Product
	json.NewDecoder(request.Body).Decode(&product)

	collection := config.NewMongoDatabase().Collection("product")

	// Save product
	result, err := collection.InsertOne(ctx, product)
	config.FailOnError(err, "Failed to insert data!")

	json.NewEncoder(response).Encode(result)
}

func Buy(response http.ResponseWriter, request *http.Request) {
	conn, err := config.GetConn()
	config.FailOnError(err, "Failed to connect RabbitMQ!")

	// Get product IDs from body request
	var ids ProductIDs
	_ = json.NewDecoder(request.Body).Decode(&ids)

	var arrIds = ids.Ids

	// Find product by ids
	products, err := GetProductsInValues("_id", arrIds)
	config.FailOnError(err, "Failed get products!")

	payload, err := json.Marshal(products)
	config.FailOnError(err, "Failed to connect to convert JSON")

	// Publish Queue
	conn.PublishQueue(payload, "ORDER")
}

func GetProductsInValues(key string, values []primitive.ObjectID) (products []Product, err error) {
	ctx, cancel := config.NewMongoContext()
	defer cancel()

	collection := config.NewMongoDatabase().Collection("product")

	filter := bson.M{key: bson.M{"$in": values}}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, err
	}
	return
}

func main() {
	log.Println("Product-Service is running!")

	router := mux.NewRouter()

	router.HandleFunc("/product/create", Create).Methods("POST")
	router.HandleFunc("/product/buy", Buy).Methods("POST")

	log.Fatal(http.ListenAndServe("localhost:8082", router))
}
