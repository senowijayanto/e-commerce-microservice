package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"product-service/config"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	SECRET_KEY = []byte("secretkey")
)

type Product struct {
	Name        string `json:"name,omitempty" bson:"name,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Price       int32  `json:"price,omitempty" bson:"price,omitempty"`
}

type ProductIDs struct {
	Ids []primitive.ObjectID `json:"ids"`
}

type Exception struct {
	Message string `json:"message"`
}

type Claims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Payload struct {
	Products []Product
	Email    string
}

type ProductWithId struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Price       int32              `json:"price,omitempty" bson:"price,omitempty"`
}

type ResponseCreate struct {
	Status  bool `json:"status"`
	Code    int  `json:"code"`
	Data    ProductWithId
	Message string `json:"message"`
}

func Create(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := config.NewMongoContext()
	defer cancel()

	response.Header().Set("Content-Type", "application/json")

	var product Product
	json.NewDecoder(request.Body).Decode(&product)

	collection := config.NewMongoDatabase().Collection("product")

	// Save product
	res, err := collection.InsertOne(ctx, product)
	config.FailOnError(err, "Failed to insert data!")
	oid := res.InsertedID.(primitive.ObjectID)

	result := ResponseCreate{
		Status: true,
		Code:   http.StatusOK,
		Data: ProductWithId{
			ID:          oid,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		},
		Message: "Your Request Has Been Processed",
	}

	json.NewEncoder(response).Encode(result)
}

func Buy(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	conn, err := config.GetConn()
	config.FailOnError(err, "Failed to connect RabbitMQ!")

	// Get product IDs from body request
	var ids ProductIDs
	_ = json.NewDecoder(request.Body).Decode(&ids)

	var unmarshal Claims

	var arrIds = ids.Ids

	claims := context.Get(request, "decoded")
	claimsMarshal, err := json.Marshal(claims)
	config.FailOnError(err, "Failed json Marshal!")

	json.Unmarshal(claimsMarshal, &unmarshal)
	userEmail := unmarshal.Email

	// Find product by ids
	products, err := GetProductsInValues("_id", arrIds)
	config.FailOnError(err, "Failed get products!")

	payload := Payload{
		Products: products,
		Email:    userEmail,
	}

	bytePayload, err := json.Marshal(payload)
	config.FailOnError(err, "Failed to convert JSON")

	// Publish Queue
	conn.PublishQueue(bytePayload, "ORDER")
	json.NewEncoder(response).Encode(payload)
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

func validateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("there was an error")
					}
					return SECRET_KEY, nil
				})
				if error != nil {
					json.NewEncoder(w).Encode(Exception{Message: error.Error()})
					return
				}
				if token.Valid {
					context.Set(req, "decoded", token.Claims)
					next(w, req)
				} else {
					json.NewEncoder(w).Encode(Exception{Message: "Invalid authorization token"})
				}
			}
		} else {
			json.NewEncoder(w).Encode(Exception{Message: "An authorization header is required"})
		}
	})
}

func main() {
	log.Println("Product-Service is running!")

	router := mux.NewRouter()

	router.HandleFunc("/product/create", validateMiddleware(Create)).Methods("POST")
	router.HandleFunc("/product/buy", validateMiddleware(Buy)).Methods("POST")

	log.Fatal(http.ListenAndServe("localhost:8082", router))
}
