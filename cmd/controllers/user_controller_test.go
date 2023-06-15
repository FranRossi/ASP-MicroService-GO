package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-service/cmd/responses"
	"user-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockClient is the mock client
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	// GetDoFunc fetches the mock client's `Do` func
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

// MockDB is a mock implementation of the database operations
type MockDB struct {
	CreateUserFunc      func(ctx context.Context, user models.UserWithCompanyAsObject) (primitive.ObjectID, error)
	FindUserByIDFunc    func(ctx context.Context, id primitive.ObjectID) (*models.UserWithCompanyAsObject, error)
	FindUserByEmailFunc func(ctx context.Context, email string) (*models.UserWithCompanyAsObject, error)
	FindAllUsersFunc    func(ctx context.Context, companyId primitive.ObjectID) ([]*models.UserWithCompanyAsObject, error)
}

// CreateUser mocks the creation of a user in the database
func (db *MockDB) CreateUser(ctx context.Context, user models.UserWithCompanyAsObject) (primitive.ObjectID, error) {
	if db.CreateUserFunc != nil {
		return db.CreateUserFunc(ctx, user)
	}
	return primitive.NilObjectID, nil
}

// FindUserByID mocks the retrieval of a user by ID from the database
func (db *MockDB) FindUserByID(ctx context.Context, id primitive.ObjectID) (*models.UserWithCompanyAsObject, error) {
	if db.FindUserByIDFunc != nil {
		return db.FindUserByIDFunc(ctx, id)
	}
	return nil, nil
}

// FindUserByEmail mocks the retrieval of a user by email from the database
func (db *MockDB) FindUserByEmail(ctx context.Context, email string) (*models.UserWithCompanyAsObject, error) {
	if db.FindUserByEmailFunc != nil {
		return db.FindUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

// FindAllUsers mocks the retrieval of all users for a company from the database
func (db *MockDB) FindAllUsers(ctx context.Context, companyId primitive.ObjectID) ([]*models.UserWithCompanyAsObject, error) {
	if db.FindAllUsersFunc != nil {
		return db.FindAllUsersFunc(ctx, companyId)
	}
	return nil, nil
}

func init() {
	Client = &MockClient{}
}

func TestCreateNewUser(t *testing.T) {
	// Create a new Gin router
	fmt.Println("Arrived TestCreateNewUser")
	router := gin.Default()

	// Set up the mock client response JSON for creating a company
	mockCompanyResponseJSON := `{
		"id": "6477c1b6f7122e3b1d204917",
		"name": "Test Company",
		"apiKey": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb21wYW55TmFtZSI6ImFiZWxpdG9oMTg1NTY1MiIsImlhdCI6MTY4NTU2OTk3NH0.tt_WYAWZnKsZQLirW3uvLgw7K56YAmvUIOd-ujJWGTE"
	}`

	// Set up the mock client response JSON for creating a user
	mockUserResponseJSON := `{
		"status": 201,
		"message": "success",
		"data": {
			"user": {
				"id": "648a26b07c0d535bb1526e1a",
				"name": "Test User",
				"password": "password",
				"email": "test@example.com",
				"role": "admin",
				"company": "Test Company"
			}
		}
	}`

	// Set up the mock client
	mockClient := &MockClient{}

	// Set up the mock response for creating a company
	mockCompanyResponseBody := []byte(mockCompanyResponseJSON)
	mockCompanyHTTPResponse := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       ioutil.NopCloser(bytes.NewReader(mockCompanyResponseBody)),
	}

	// Set up the mock response for creating a user
	mockUserResponseBody := []byte(mockUserResponseJSON)
	mockUserHTTPResponse := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       ioutil.NopCloser(bytes.NewReader(mockUserResponseBody)),
	}

	// Set up the DoFunc for the mock client
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/companies" {
			// If the request is for creating a company, return the mock company response
			return mockCompanyHTTPResponse, nil
		} else if req.URL.Path == "/users" {
			// If the request is for creating a user, return the mock user response
			return mockUserHTTPResponse, nil
		}
		return nil, fmt.Errorf("unexpected request path: %s", req.URL.Path)
	}

	// Assign the mock client's DoFunc to the GetDoFunc variable
	GetDoFunc = mockClient.DoFunc

	// Assign the mock client to the controller
	Client = mockClient

	mockDB := &MockDB{}
	mockDB.CreateUserFunc = func(ctx context.Context, user models.UserWithCompanyAsObject) (primitive.ObjectID, error) {
		id, _ := primitive.ObjectIDFromHex("648a26b07c0d535bb1526e1a")
		return id, nil
	}

	DB = mockDB

	// Set up the route
	router.POST("/users", CreateUser())

	// Create a custom request payload
	requestPayload := models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password",
		Role:     "admin",
		Company:  "Test Company",
	}

	// Convert the request payload to JSON
	payload, _ := json.Marshal(requestPayload)

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

	// Assert specific fields in the response data
	userData := response.Data["user"].(map[string]interface{})
	assert.Equal(t, "648a26b07c0d535bb1526e1a", userData["_id"])
	assert.Equal(t, "Test User", userData["name"])
	assert.Equal(t, "password", userData["password"])
	assert.Equal(t, "test@example.com", userData["email"])
	assert.Equal(t, "admin", userData["role"])
	assert.Equal(t, "Test Company", userData["company"])
}

func TestCreateUserInvalidUserName(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the route
	router.POST("/users", CreateUser())

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
