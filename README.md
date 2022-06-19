# E-Commerce Microservices
This project is a simple skeleton code for microservice architecture pattern using Golang and RabbitMQ.

## Diagram
![Microservices](./microservice-diagram.png?raw=true "Microservices Diagram")

## Requirements
* [Docker](https://docs.docker.com/get-docker/)
* [Docker Compose](https://docs.docker.com/compose/install/)

## Installation
1. Clone this git repoistory.

    ```bash
    https://github.com/senowijayanto/e-commerce-microservice
    ```

2. Go to the e-commerce-microservice directory.

    ```bash
    cd e-commerce-microservice
    ```

3. Start the docker composed cluster.

    ```bash
    docker-compose up -d
    ```

3. Running from local.

    ```bash
    http://localhost:8080/api/v1/users
    http://localhost:8080/api/v1/products
    http://localhost:8080/api/v1/orders
    ```

## Services
This project was decomposed into three cores microservices. All of them are independently deployable applications, organized around certain business domains.

### Auth Service
Provides several API for Auth Service.
| Method | Path                   | Description               |
|--------|------------------------|---------------------------|
| POST   | /auth/register         | Create new user           |
| POST   | /auth/login            | Login to get access Token |


**_User Registration_**
* Request
```
Path         : http://localhost:8081/auth/register
Method       : POST
Content-Type : application/json
Body         :
{
    "name"    : "Adam Smith"
    "email"   : "adam.smith@gmail.com",
    "password": "adamsecret"
}
```
* Response
```
Status  : 200
{
	"id"    : "asd21312412",
	"name"  : "Adam Smith",
	"email" : "adam.smith@gmail.com"
}
```
**_User Login_**
* Request
```
Path         : http://localhost:8081/auth/login
Method       : POST
Content-Type : application/json
Body :
{
    "email": "adam.smith@gmail.com",
    "password": "adamsecret"
}
```
* Response
```
Status  : 200
{
	"message" : "Login Success!",
	"token"   : "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InNlbm93aWpheWFudG9AZ21haWwuY29tIiwibmFtZSI6IlNlbm8iLCJuYmYiOjE0NDQ0Nzg0MDB9.cTlXs8yXH-QxSroK0SZ5pQ7sLSzFrTYr9hJqk1NXCuA"
}
```


### Product Service
Provides several API for Product Service.
| Method | Path                 | Description                  |
|--------|----------------------|------------------------------|
| POST   | /product/create      | Create new product           |
| POST   | /product/buy         | Buy some product from Orders |



### Order Service


