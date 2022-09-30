package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"auth-service/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	SECRET_KEY = []byte("secretkey")
)

type User struct {
	Id       string `json:"id" bson:"id"`
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}

type UserResponse struct {
	Id    primitive.ObjectID `json:"id" bson:"id"`
	Name  string             `json:"name" bson:"name"`
	Email string             `json:"email" bson:"email"`
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

type UserToken struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type ResponseStatus struct {
	Status  bool `json:"status"`
	Code    int  `json:"code"`
	Data    UserResponse
	Message string `json:"message"`
}

type TokenResponse struct {
	Status  bool `json:"status"`
	Code    int  `json:"code"`
	Data    UserToken
	Message string `json:"message"`
}

func Register(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := config.NewMongoContext()
	defer cancel()

	response.Header().Set("Content-Type", "application/json")

	var user User
	var dbUser User
	json.NewDecoder(request.Body).Decode(&user)

	user.Password = GetHash([]byte(user.Password))

	collection := config.NewMongoDatabase().Collection("user")

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
	oid := res.InsertedID.(primitive.ObjectID)

	result := ResponseStatus{
		Status: true,
		Code:   http.StatusOK,
		Data: UserResponse{
			Id:    oid,
			Name:  user.Name,
			Email: user.Email,
		},
		Message: "Your Request Has Been Processed",
	}

	json.NewEncoder(response).Encode(result)
}

func Login(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := config.NewMongoContext()
	defer cancel()

	response.Header().Set("Content-Type", "application/json")

	var user User
	var dbUser User
	json.NewDecoder(request.Body).Decode(&user)

	collection := config.NewMongoDatabase().Collection("user")

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

	result := TokenResponse{
		Status: true,
		Code:   http.StatusOK,
		Data: UserToken{
			Email: user.Email,
			Token: tokenString,
		},
		Message: "Your Request Has Been Processed",
	}

	json.NewEncoder(response).Encode(result)
}

func main() {
	log.Println("Auth-Service is running!")

	router := mux.NewRouter()

	router.HandleFunc("/auth/register", Register).Methods("POST")
	router.HandleFunc("/auth/login", Login).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
