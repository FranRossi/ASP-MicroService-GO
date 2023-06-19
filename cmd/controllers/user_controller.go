package controllers

import (
	"context"
	"net/http"
	"time"
	"user-service/cmd/responses"
	"user-service/internal/configs"
	"user-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HTTPClient
	DB     configs.Database
)

var validate = validator.New()

func init() {
	Client = &http.Client{}
	DB = configs.NewMongoDB(configs.ConnectDB())
}

func CreateUser() gin.HandlerFunc {
	log.Info().Msg("Create user endpoint reached")
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var user models.User
		defer cancel()
		log.Info().Msg("User being created:" + user.Password)
		if err := c.BindJSON(&user); err != nil {
			log.Error().Err(err).Msg("error wrong json format")
			return
		}

		err := ValidateRequest(user, c)
		if err != nil {
			log.Error().Err(err).Msg("Error validating request")
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "validation error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		if user, _ := DB.FindUserByEmail(ctx, user.Email); user != nil {
			log.Error().Msg("User already exists with email: " + user.Email)
			c.JSON(http.StatusBadRequest, responses.UserResponse{
				Status:  http.StatusBadRequest,
				Message: "User already exists with email: " + user.Email,
				Data: map[string]interface{}{
					"data": "User already exists with email: " + user.Email,
				},
			})
			return
		}

		// var companyIdObject primitive.ObjectID
		// var err2 error = nil

		// if user.Invitation {
		// 	log.Info().Msg("user is invited")
		// 	companyIdObject, err2 = primitive.ObjectIDFromHex(user.Company)
		// } else {
		// 	companyIdStr, err := CreateCompany(user.Company)
		// 	if err != nil {
		// 		log.Error().Err(err).Msg("Error creating company")
		// 		c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "company error", Data: map[string]interface{}{"data": err.Error()}})
		// 		return
		// 	}

		// 	companyIdObject, err2 = primitive.ObjectIDFromHex(companyIdStr)
		// }
		companyIdObject, err2 := primitive.ObjectIDFromHex(user.Company)

		if err2 != nil {
			log.Error().Err(err2).Msg("Error converting company ID to object")
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error on companyId as an object", Data: map[string]interface{}{"data": err2.Error()}})
			return
		}
		userWithCompany := models.UserWithCompanyAsObject{
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
			Role:     user.Role,
			Company:  companyIdObject,
		}

		// Call the CreateUser method on the DB interface
		userId, err := DB.CreateUser(ctx, userWithCompany)
		if err != nil {
			log.Error().Err(err).Msg("Error storing a user on database")
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error storing user on database", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		log.Info().Msg("User created successfully")
		user.Id = userId.Hex()

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"user": user}})
	}
}

func ValidateRequest(user models.User, c *gin.Context) error {

	//use the validator library to validate required fields
	if validationErr := validate.Struct(&user); validationErr != nil {
		log.Error().Err(validationErr).Msg("error validating request fields")
		return validationErr
	}

	return nil
}

func FindById() gin.HandlerFunc {
	log.Info().Msg("Get a specific user endpoint by id reached")
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(userId)

		userWithCompany, err := DB.FindUserByID(ctx, objId)
		if err != nil {
			log.Error().Err(err).Msg("Error getting a user from database")
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error getting a user from database", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		log.Info().Msg("User: " + userId + " retrieved successfully")
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"user": userWithCompany}})
	}
}

// func CreateCompany(company string) (string, error) {
// 	url := os.Getenv("COMPANY_SERVICE_URL")
// 	jsonStr := []byte(`{"name":"` + company + `"}`)

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err := Client.Do(req)

// 	if err != nil || resp.StatusCode != http.StatusCreated {
// 		log.Error().Err(err).Msg("Error creating a new company on separate service")
// 		return "", errors.New("failed creating a new company on separate service")
// 	}
// 	defer resp.Body.Close()
// 	bodyBytes, err := ioutil.ReadAll(resp.Body)
// 	if err != nil || len(bodyBytes) == 0 {
// 		log.Error().Err(err).Msg("Error reading response body when creating a new company on separate service")
// 		return "", errors.New("failed reading response body when creating a new company on separate service")
// 	}

// 	// Extract values from response body
// 	var responseMap map[string]interface{}
// 	_ = json.Unmarshal(bodyBytes, &responseMap)

// 	companyIdStr, ok := responseMap["id"].(string)
// 	if !ok {
// 		log.Error().Msg("Fail to extract companny ID")
// 		return "", errors.New("failed to extract company Id from response body when creating a new company on separate service")
// 	}
// 	log.Info().Msg("Company created on separate service successfully")

// 	return companyIdStr, nil
// }

func GetUsers() gin.HandlerFunc {
	log.Info().Msg("Get all users endpoint reached")
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		email := c.Query("email")
		if email != "" {
			FindByEmail(c, email)
			return
		}

		companyId := c.Query("company")
		if companyId == "" || companyId == "undefined" {
			log.Error().Msg("Error getting user for a company, Company query parameter is missing")
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "Error getting user for a company, Company query parameter is missing", Data: nil})
			return
		}

		objId, _ := primitive.ObjectIDFromHex(companyId)
		usersList, err := DB.FindAllUsers(ctx, objId)
		if err != nil {
			log.Error().Err(err).Msg("There was a problem trying to find users on database with this compnay Id: " + companyId)
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "There was a problem trying to find users on database", Data: nil})
			return
		}

		log.Info().Msg("Users retrieved successfully!")
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"users": usersList}})
	}
}

func FindByEmail(c *gin.Context, email string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info().Msg("Looking for user: " + email)
	userWithCompany, err := DB.FindUserByEmail(ctx, email)
	if err != nil {
		log.Error().Err(err).Msg("Error getting a user from database with email: " + email)
		c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error getting a user from database with provided email", Data: map[string]interface{}{"data": err.Error()}})
		return
	}

	log.Info().Msg("User: " + email + " retrieved successfully")
	c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"user": userWithCompany}})
}
