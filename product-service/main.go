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

type ProductPayload struct {
	Id          string `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string `json:"name,omitempty" bson:"name,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Price       int32  `json:"price,omitempty" bson:"price,omitempty"`
}

var (
	SECRET_KEY                 = []byte("secretkey")
	MONGO_URI                  = "mongodb://admin:secret@localhost:27019"
	MONGO_DATABASE             = "productDB"
	MONGO_POOL_MIN             = 10
	MONGO_POOL_MAX             = 100
	MONGO_MAX_IDLE_TIME_SECOND = 60
)

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
	ctx, cancel := config.NewMongoContext()
	defer cancel()

	// Get product from body request
	var product Product
	var dbProduct Product
	_ = json.NewDecoder(request.Body).Decode(&product)

	// Find product by id
	collection := config.NewMongoDatabase().Collection("product")
	collection.FindOne(ctx, bson.M{"_id": product.Id}).Decode(&dbProduct)

	bodyProduct := ProductPayload{
		Id:          primitive.ObjectID(dbProduct.Id).Hex(),
		Name:        dbProduct.Name,
		Description: dbProduct.Description,
		Price:       dbProduct.Price,
	}
	payload, err := json.Marshal(bodyProduct)
	config.FailOnError(err, "Failed to connect to convert JSON")

	// Publish Queue
	config.PublishQueue(payload, "ORDER")

	// return to json
	config.ConsumeQueue("PRODUCT")

}

func main() {
	log.Println("Product-Service is running!")

	router := mux.NewRouter()

	router.HandleFunc("/product/create", Create).Methods("POST")
	router.HandleFunc("/product/buy", Buy).Methods("POST")

	log.Fatal(http.ListenAndServe("localhost:8082", router))
}
