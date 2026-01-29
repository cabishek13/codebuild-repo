package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

type Employee struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Address  string `json:"address"`
	IsActive *bool  `json:"is_active"`
}

var db *sql.DB

func init() {
	var err error

	dsn := "intern:UclYkD0jzxpLUvYl@tcp(db-interns26.cgsudm7fyigu.us-west-2.rds.amazonaws.com:3306)/development"

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Error opening DB: %v\n", err)
	}
	fmt.Printf("Successfully connected to database\n")

	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if err := db.Ping(); err != nil {
		fmt.Printf("Database Connection Failed: %v\n", err)
		return apiResponse(http.StatusInternalServerError, "Error connecting to database"), nil
	}

	switch req.HTTPMethod {
	case "GET":
		fmt.Println("Request for GET /employees received")
		return handleGet(req), nil
	case "POST":
		fmt.Println("Request for POST /employees received")
		return handlePost(req)
	default:
		return apiResponse(http.StatusMethodNotAllowed, "Method Not Allowed"), nil
	}
}

func handleGet(req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	val, hasFilter := req.QueryStringParameters["is_active"]

	var query string
	var rows *sql.Rows
	var err error

	if hasFilter && val != "" {
		wanted := (val == "true")
		query = "SELECT id, name, age, address, is_active FROM cabishek_employees WHERE is_active = ?"
		rows, err = db.Query(query, wanted)
	} else {
		query = "SELECT id, name, age, address, is_active FROM cabishek_employees"
		rows, err = db.Query(query)
	}

	if err != nil {
		fmt.Printf("Query Error: %v\n", err)
		return apiResponse(http.StatusInternalServerError, "Error while accessing data resource.")
	}
	defer rows.Close()

	var results []Employee
	for rows.Next() {
		var e Employee
		var isActive bool

		if err := rows.Scan(&e.ID, &e.Name, &e.Age, &e.Address, &isActive); err != nil {
			fmt.Printf("Scan Error: %v\n", err)
			continue
		}
		e.IsActive = &isActive
		results = append(results, e)
	}
	if results == nil {
		results = []Employee{}
	}
	fmt.Println("Request GET /employees was processed successfully")

	return jsonResponse(http.StatusOK, results)
}

func handlePost(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var newEmp Employee

	if err := json.Unmarshal([]byte(req.Body), &newEmp); err != nil {
		return apiResponse(http.StatusBadRequest, "Invalid JSON format"), nil
	}

	if newEmp.Name == "" || newEmp.Age == 0 || newEmp.Address == "" {
		return apiResponse(http.StatusBadRequest, "Missing required fields: name, age, and address are required"), nil
	}

	newEmp.Name = strings.ToLower(newEmp.Name)
	newEmp.Address = strings.ToLower(newEmp.Address)
	if newEmp.IsActive == nil {
		t := true
		newEmp.IsActive = &t
	}
	var exists int
	err := db.QueryRow("SELECT 1 FROM cabishek_employees WHERE name = ?", newEmp.Name).Scan(&exists)
	if err == nil {
		return apiResponse(http.StatusConflict, "Data resource already exists. Use PUT to update data."), nil
	} else if err != sql.ErrNoRows {
		fmt.Printf("Check Error: %v\n", err)
		return apiResponse(http.StatusInternalServerError, "Error while accessing data resource."), nil
	}

	query := "INSERT INTO cabishek_employees (name, age, address, is_active) VALUES (?, ?, ?, ?)"
	_, err = db.Exec(query, newEmp.Name, newEmp.Age, newEmp.Address, *newEmp.IsActive)

	if err != nil {
		fmt.Printf("Insert Error: %v\n", err)
		return apiResponse(http.StatusInternalServerError, "Error while accessing data resource."), nil
	}
	fmt.Println("Request POST /employees was processed successfully")
	return apiResponse(http.StatusOK, "Data resource created successfully."), nil
}

func apiResponse(status int, message string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       message,
		Headers:    map[string]string{"Content-Type": "text/plain"},
	}
}

func jsonResponse(status int, data interface{}) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(data)
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}
