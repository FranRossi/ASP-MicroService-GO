package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
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
	router := gin.Default()

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
				"company": "649060d540e3b169621e9629"
			}
		}
	}`

	// Set up the mock client
	mockClient := &MockClient{}

	// Set up the mock response for creating a user
	mockUserResponseBody := []byte(mockUserResponseJSON)
	mockUserHTTPResponse := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       ioutil.NopCloser(bytes.NewReader(mockUserResponseBody)),
	}

	// Set up the DoFunc for the mock client
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/users" {
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
	mockDB.FindUserByEmailFunc = func(ctx context.Context, email string) (*models.UserWithCompanyAsObject, error) {
		return nil, nil
	}
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
		Company:  "649060d540e3b169621e9629",
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
	assert.Equal(t, "649060d540e3b169621e9629", userData["company"])
}

func TestUserAlreadyExist(t *testing.T) {
	router := gin.Default()

	user := models.UserWithCompanyAsObject{
		Id:       primitive.NewObjectID(),
		Name:     "Test User",
		Email:    "test@gmail.com",
		Password: "password",
		Role:     "admin",
		Company:  primitive.NewObjectID(),
	}
	mockDB := &MockDB{}
	mockDB.FindUserByEmailFunc = func(ctx context.Context, email string) (*models.UserWithCompanyAsObject, error) {

		return &user, nil
	}

	// Set up the mock client response JSON for creating a user
	mockUserResponseJSON := `{
		"status": 400,
		"message": "User already exists with email: ` + user.Email + `",
		"data": {
			"data": "User already exists with email: fran@gmail.com"
		}
	}`

	// Set up the mock client
	mockClient := &MockClient{}

	// Set up the mock response for creating a user
	mockUserResponseBody := []byte(mockUserResponseJSON)
	mockUserHTTPResponse := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(bytes.NewReader(mockUserResponseBody)),
	}

	// Set up the DoFunc for the mock client
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/users" {
			// If the request is for creating a user, return the mock user response
			return mockUserHTTPResponse, nil
		}
		return nil, fmt.Errorf("unexpected request path: %s", req.URL.Path)
	}

	// Assign the mock client's DoFunc to the GetDoFunc variable
	GetDoFunc = mockClient.DoFunc

	// Assign the mock client to the controller
	Client = mockClient

	mockDB.FindUserByEmailFunc = func(ctx context.Context, email string) (*models.UserWithCompanyAsObject, error) {
		return &user, nil
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
		Company:  user.Company.Hex(),
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
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// Parse the response body
	var response responses.UserResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check the response message
	assert.Equal(t, "User already exists with email: "+user.Email, response.Message)
}

func TestCreateUserMissingUserNameField(t *testing.T) {
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

func TestGetUserByID(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the mock database response
	mockUserID := primitive.NewObjectID()
	mockUser := &models.UserWithCompanyAsObject{
		Id:       mockUserID,
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password",
		Role:     "admin",
		Company:  primitive.NewObjectID(),
	}

	// Set up the mock database
	mockDB := &MockDB{
		FindUserByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*models.UserWithCompanyAsObject, error) {
			assert.Equal(t, mockUserID, id, "expected user ID to match")
			return mockUser, nil
		},
	}

	// Set up the controller with the mock DB
	DB = mockDB

	// Define the expected response
	expectedResponse := responses.UserResponse{
		Status:  http.StatusOK,
		Message: "success",
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"_id":      mockUser.Id.Hex(),
				"name":     mockUser.Name,
				"email":    mockUser.Email,
				"password": mockUser.Password,
				"role":     mockUser.Role,
				"company":  mockUser.Company.Hex(),
			},
		},
	}

	// Define the test request
	requestURL := fmt.Sprintf("/users/%s", mockUserID.Hex())
	request, _ := http.NewRequest("GET", requestURL, nil)

	// Perform the request
	responseRecorder := httptest.NewRecorder()
	router.GET("/users/:userId", FindById())
	router.ServeHTTP(responseRecorder, request)

	// Validate the response
	assert.Equal(t, http.StatusOK, responseRecorder.Code, "expected status OK")
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	var actualResponse responses.UserResponse
	_ = json.Unmarshal(responseBody, &actualResponse)
	assert.Equal(t, expectedResponse, actualResponse, "expected response to match")
}

func TestFindByEmail(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the mock database response
	mockUser := &models.UserWithCompanyAsObject{
		Id:       primitive.NewObjectID(),
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password",
		Role:     "admin",
		Company:  primitive.NewObjectID(),
	}

	// Set up the mock database
	mockDB := &MockDB{
		FindUserByEmailFunc: func(ctx context.Context, email string) (*models.UserWithCompanyAsObject, error) {
			assert.Equal(t, mockUser.Email, email, "expected email to match")
			return mockUser, nil
		},
	}

	// Set up the controller with the mock DB
	DB = mockDB

	// Define the expected response
	expectedResponse := responses.UserResponse{
		Status:  http.StatusOK,
		Message: "success",
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"_id":      mockUser.Id.Hex(),
				"name":     mockUser.Name,
				"email":    mockUser.Email,
				"password": mockUser.Password,
				"role":     mockUser.Role,
				"company":  mockUser.Company.Hex(),
			},
		},
	}

	// Define the test request
	requestURL := "/users?email=" + mockUser.Email
	request, _ := http.NewRequest("GET", requestURL, nil)

	// Perform the request
	responseRecorder := httptest.NewRecorder()
	router.GET("/users", func(c *gin.Context) {
		FindByEmail(c, c.Query("email"))
	})
	router.ServeHTTP(responseRecorder, request)

	// Validate the response
	assert.Equal(t, http.StatusOK, responseRecorder.Code, "expected status OK")
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	var actualResponse responses.UserResponse
	_ = json.Unmarshal(responseBody, &actualResponse)
	assert.Equal(t, expectedResponse, actualResponse, "expected response to match")
}

func TestGetUsers(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the mock database response
	mockUsers := []*models.UserWithCompanyAsObject{
		{
			Id:       primitive.NewObjectID(),
			Name:     "John Doe",
			Email:    "john.doe@example.com",
			Password: "password",
			Role:     "admin",
			Company:  primitive.NewObjectID(),
		},
		{
			Id:       primitive.NewObjectID(),
			Name:     "Jane Smith",
			Email:    "jane.smith@example.com",
			Password: "password",
			Role:     "user",
			Company:  primitive.NewObjectID(),
		},
	}

	// Set up the mock database
	mockDB := &MockDB{
		FindAllUsersFunc: func(ctx context.Context, companyID primitive.ObjectID) ([]*models.UserWithCompanyAsObject, error) {
			// You can perform additional assertions here if needed
			return mockUsers, nil
		},
	}

	// Set up the controller with the mock DB
	DB = mockDB

	// Define the expected response
	expectedResponse := responses.UserResponse{
		Status:  http.StatusOK,
		Message: "success",
		Data: map[string]interface{}{
			"users": mockUsers,
		},
	}

	// Define the test request
	requestURL := "/users?company=" + mockUsers[0].Company.Hex()
	request, _ := http.NewRequest("GET", requestURL, nil)

	// Perform the request
	responseRecorder := httptest.NewRecorder()
	router.GET("/users", GetUsers())
	router.ServeHTTP(responseRecorder, request)

	// Validate the response
	assert.Equal(t, http.StatusOK, responseRecorder.Code, "expected status OK")
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	var actualResponse responses.UserResponse
	_ = json.Unmarshal(responseBody, &actualResponse)

	// Assert the expected response status and message
	assert.Equal(t, expectedResponse.Status, actualResponse.Status, "expected response status to match")
	assert.Equal(t, expectedResponse.Message, actualResponse.Message, "expected response message to match")

	// Assert the expected response data
	actualUsers := actualResponse.Data["users"].([]interface{})
	assert.Equal(t, len(mockUsers), len(actualUsers), "expected number of users to match")

	for i, user := range mockUsers {
		actualUser := actualUsers[i].(map[string]interface{})
		assert.Equal(t, user.Id.Hex(), actualUser["_id"].(string), "expected user ID to match")
		assert.Equal(t, user.Name, actualUser["name"].(string), "expected user name to match")
		assert.Equal(t, user.Email, actualUser["email"].(string), "expected user email to match")
		assert.Equal(t, user.Password, actualUser["password"].(string), "expected user password to match")
		assert.Equal(t, user.Role, actualUser["role"].(string), "expected user role to match")
	}
}

func TestCreateUserBindJsonError(t *testing.T) {
	// Create a mock request with an invalid JSON format
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to capture the response
	w := httptest.NewRecorder()

	// Create a Gin context with the mock request and response
	context, _ := gin.CreateTestContext(w)
	context.Request = req

	// Call the CreateUser middleware function
	CreateUser()(context)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvitewUser(t *testing.T) {
	router := gin.Default()

	// Set up the mock client response JSON for creating a company
	mockCompanyResponseJSON := `{
		"id": "606d97b4c1bea43ce49be6dc",
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
				"company": "606d97b4c1bea43ce49be6dc"
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
		Company:  "606d97b4c1bea43ce49be6dc", // Replace with a valid company ID
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
	assert.Equal(t, "606d97b4c1bea43ce49be6dc", userData["company"])
}

func TestErrorWrongObjectId(t *testing.T) {
	router := gin.Default()

	// Set up the mock client response JSON for creating a user
	mockUserResponseJSON := `{
		"status": 500,
		"message": "error on companyId as an object",
		"data": {
			"error": "companyId must be a string"
		}
	}`

	// Set up the mock client
	mockClient := &MockClient{}

	// Set up the mock response for creating a user
	mockUserResponseBody := []byte(mockUserResponseJSON)
	mockUserHTTPResponse := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(bytes.NewReader(mockUserResponseBody)),
	}

	// Set up the DoFunc for the mock client
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/users" {
			// If the request is for creating a company, return the mock company response
			return mockUserHTTPResponse, nil
		}
		return nil, fmt.Errorf("unexpected request path: %s", req.URL.Path)
	}

	// Assign the mock client's DoFunc to the GetDoFunc variable
	GetDoFunc = mockClient.DoFunc

	// Assign the mock client to the controller
	Client = mockClient

	// Set up the route
	router.POST("/users", CreateUser())

	// Create a custom request payload
	requestPayload := models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password",
		Role:     "admin",
		Company:  "606d97b4c1bea43ce49be6dc_!WorngId",
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
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	// Parse the response body
	var response responses.UserResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check the response message
	assert.Equal(t, "error on companyId as an object", response.Message)
}

func TestErrorDatabase(t *testing.T) {
	router := gin.Default()

	// Set up the mock client response JSON for creating a company
	mockCompanyResponseJSON := `{
		"id": "606d97b4c1bea43ce49be6dc",
		"name": "Test Company",
		"apiKey": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb21wYW55TmFtZSI6ImFiZWxpdG9oMTg1NTY1MiIsImlhdCI6MTY4NTU2OTk3NH0.tt_WYAWZnKsZQLirW3uvLgw7K56YAmvUIOd-ujJWGTE"
	}`

	// Set up the mock client response JSON for creating a user
	mockUserResponseJSON := `{
		"status": 500,
		"message": "error storing user on database",
		"data": {
			"error": "error storing user on database"
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
		StatusCode: http.StatusInternalServerError,
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
		return primitive.ObjectID{}, fmt.Errorf("error storing user on database")
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
		Company:  "606d97b4c1bea43ce49be6dc",
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
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	// Parse the response body
	var response responses.UserResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check the response message
	assert.Equal(t, "error storing user on database", response.Message)
}

func TestErrorFindUserByID(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the mock client
	mockClient := &MockClient{}
	// ...

	// Assign the mock client to the controller
	Client = mockClient

	mockDB := &MockDB{}
	mockDB.FindUserByIDFunc = func(ctx context.Context, id primitive.ObjectID) (*models.UserWithCompanyAsObject, error) {
		return nil, errors.New("mock find user by ID error")
	}
	DB = mockDB

	// Set up the route
	router.GET("/users/:userId", FindById())

	// Create a GET request
	req, _ := http.NewRequest("GET", "/users/123", nil)

	// Perform the request and record the response
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Check the response status code
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	// Parse the response body
	var response responses.UserResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check the response message
	assert.Equal(t, "Error getting a user from database", response.Message)

	// Assert the error message in the response data
	assert.Equal(t, "mock find user by ID error", response.Data["data"])
}

func TestErrorFindUserByEmail(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the mock client
	mockClient := &MockClient{}
	// ...

	// Assign the mock client to the controller
	Client = mockClient

	mockDB := &MockDB{}
	mockDB.FindUserByEmailFunc = func(ctx context.Context, email string) (*models.UserWithCompanyAsObject, error) {
		return nil, errors.New("mock find user by email error")
	}
	DB = mockDB

	// Set up the route
	router.GET("/users", GetUsers())

	// Create a GET request
	req, _ := http.NewRequest("GET", "/users?email=test@gmail.com", nil)

	// Perform the request and record the response
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Check the response status code
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	// Parse the response body
	var response responses.UserResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check the response message
	assert.Equal(t, "Error getting a user from database with provided email", response.Message)

	// Assert the error message in the response data
	assert.Equal(t, "mock find user by email error", response.Data["data"])
}

func TestErrorFindUsersEmptyCompanyId(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the mock client
	mockClient := &MockClient{}
	// ...

	// Assign the mock client to the controller
	Client = mockClient

	// Set up the route
	router.GET("/users", GetUsers())

	// Create a GET request
	req, _ := http.NewRequest("GET", "/users?company=", nil)

	// Perform the request and record the response
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Check the response status code
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// Parse the response body
	var response responses.UserResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check the response message
	assert.Equal(t, "Error getting user for a company, Company query parameter is missing", response.Message)
}

func TestErrorFindUsersDatabase(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the mock client
	mockClient := &MockClient{}
	// ...

	// Assign the mock client to the controller
	Client = mockClient

	mockDB := &MockDB{}
	mockDB.FindAllUsersFunc = func(ctx context.Context, companyId primitive.ObjectID) ([]*models.UserWithCompanyAsObject, error) {
		return nil, errors.New("mock find all users error")
	}
	DB = mockDB

	// Set up the route
	router.GET("/users", GetUsers())

	// Create a GET request
	req, _ := http.NewRequest("GET", "/users?company=64dc", nil)

	// Perform the request and record the response
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Check the response status code
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	// Parse the response body
	var response responses.UserResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check the response message
	assert.Equal(t, "There was a problem trying to find users on database", response.Message)
}
