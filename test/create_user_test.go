package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-service/cmd/controllers"
	"user-service/cmd/responses"
	"user-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockCreateCompany struct{}

func (m *MockCreateCompany) CreateCompany(company string) (string, error) {
	return "mockedCompanyID", nil
}

func TestCreateUserWithMockCreateCompany(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the route
	// mockCreateCompany := &MockCreateCompany{}
	router.POST("/users", controllers.CreateUser())

	// Create a sample user payload
	user := models.User{
		Name:       "John Doe",
		Email:      "john.doe@example.com",
		Password:   "password",
		Role:       "user",
		Company:    "Company XYZ",
		Invitation: false,
	}

	// Convert user struct to JSON
	payload, _ := json.Marshal(user)

	// Create a POST request with the payload
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request and record the response
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Check the response status code
	assert.Equal(t, http.StatusCreated, resp.Code)

	// Parse the response body
	var response responses.UserResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check the response message
	assert.Equal(t, "success", response.Message)

	// Assert that the user ID is not empty
	assert.NotEmpty(t, response.Data["_id"])
}

func TestCreateUserInvalidUserName(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the route
	router.POST("/users", controllers.CreateUser())

	// Create a sample user payload with no name
	user := models.User{
		Email:    "john.doe@example.com",
		Password: "password",
		Role:     "user",
		Company:  "Company XYZ",
	}

	// Convert user struct to JSON
	payload, _ := json.Marshal(user)

	// Create a POST request with the payload
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request and record the response
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Check the response status code
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// Parse the response body
	var response responses.UserResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check the response message
	assert.Equal(t, "validation error", response.Message)
}
