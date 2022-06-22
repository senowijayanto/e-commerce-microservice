package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	SECRET_KEY                 = []byte("secretkey")
	MONGO_URI                  = "mongodb://admin:secret@localhost:27018"
	MONGO_DATABASE             = "authDB"
	MONGO_POOL_MIN             = 10
	MONGO_POOL_MAX             = 100
	MONGO_MAX_IDLE_TIME_SECOND = 60
)

type User struct {
	Id       string `json:"id" bson:"id"`
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}

type UserResponse struct {
	Id    interface{} `json:"id" bson:"id"`
	Name  string      `json:"name" bson:"name"`
	Email string      `json:"email" bson:"email"`
}

type MyClaims struct {
	jwt.StandardClaims
	Name  string `json:"name"`
	Email string `json:"email"`
}

func GetHash(password []byte) string {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
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

func Register(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := NewMongoContext()
	defer cancel()

	response.Header().Set("Content-Type", "application/json")

	var user User
	var dbUser User
	json.NewDecoder(request.Body).Decode(&user)

	user.Password = GetHash([]byte(user.Password))

	collection := NewMongoDatabase().Collection("user")

	// Check existing user by an email
	collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)
	if dbUser.Email != "" {
		response.WriteHeader(http.StatusConflict)
		result := map[string]string{
			"message": "User with email " + dbUser.Email + " is already exist",
		}
		json.NewEncoder(response).Encode(result)
		return
	}

	// Save user
	res, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Panic(err)
	}

	result := UserResponse{
		Id:    res.InsertedID,
		Name:  user.Name,
		Email: user.Email,
	}

	json.NewEncoder(response).Encode(result)
}

func Login(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := NewMongoContext()
	defer cancel()

	response.Header().Set("Content-Type", "application/json")

	var user User
	var dbUser User
	json.NewDecoder(request.Body).Decode(&user)

	collection := NewMongoDatabase().Collection("user")

	// Check existing user by an email
	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		result := map[string]string{
			"message": err.Error(),
		}
		json.NewEncoder(response).Encode(result)
		return
	}

	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	// Check password
	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

	if passErr != nil {
		response.WriteHeader(http.StatusInternalServerError)
		result := map[string]string{
			"message": "Wrong Password!",
		}
		json.NewEncoder(response).Encode(result)
		return
	}

	// Generate Token
	claims := MyClaims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "Auth Service",
			ExpiresAt: time.Now().Add(time.Duration(5) * time.Minute).Unix(),
		},
		Name:  dbUser.Name,
		Email: dbUser.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(SECRET_KEY)

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		result := map[string]string{
			"message": err.Error(),
		}
		json.NewEncoder(response).Encode(result)
		return
	}

	result := map[string]string{
		"message": "Login Success!",
		"token":   tokenString,
	}

	json.NewEncoder(response).Encode(result)
}

func main() {
	log.Println("Auth-Service is running!")

	router := mux.NewRouter()

	router.HandleFunc("/auth/register", Register).Methods("POST")
	router.HandleFunc("/auth/login", Login).Methods("POST")

	log.Fatal(http.ListenAndServe("localhost:8081", router))
}
