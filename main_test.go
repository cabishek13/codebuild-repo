package main

import (
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandleGet(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
	}

	resp, err := handleRequest(req)
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	fmt.Println("--- GET Response Body ---")
	fmt.Println(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandlePost(t *testing.T) {
	payload := `{"id": 102, "name": "test_user_local1", "age": 26, "address": "local_host1"}`

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Body:       payload,
	}
	resp, err := handleRequest(req)
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	fmt.Println("--- POST Response Body ---")
	fmt.Println(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		t.Errorf("Expected status 200/201, got %d. Error: %s", resp.StatusCode, resp.Body)
	}
}
