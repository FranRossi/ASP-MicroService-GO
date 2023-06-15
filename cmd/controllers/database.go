package controllers

import (
	"context"
	"user-service/internal/configs"
	"user-service/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Database interface
type Database interface {
	CreateUser(ctx context.Context, user models.UserWithCompanyAsObject) (primitive.ObjectID, error)
	FindUserByID(ctx context.Context, id primitive.ObjectID) (*models.UserWithCompanyAsObject, error)
	FindUserByEmail(ctx context.Context, email string) (*models.UserWithCompanyAsObject, error)
	FindAllUsers(ctx context.Context, companyId primitive.ObjectID) ([]*models.UserWithCompanyAsObject, error)
}

// MongoDB implements the Database interface
type MongoDB struct {
	client         *mongo.Client
	userCollection *mongo.Collection
}

// NewMongoDB creates a new MongoDB instance
func NewMongoDB(client *mongo.Client) *MongoDB {
	return &MongoDB{
		client:         client,
		userCollection: configs.GetCollection(client, "users"),
	}
}

// CreateUser creates a new user in the database
func (db *MongoDB) CreateUser(ctx context.Context, user models.UserWithCompanyAsObject) (primitive.ObjectID, error) {
	result, err := db.userCollection.InsertOne(ctx, user)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (db *MongoDB) FindUserByID(ctx context.Context, id primitive.ObjectID) (*models.UserWithCompanyAsObject, error) {
	var user models.UserWithCompanyAsObject
	err := db.userCollection.FindOne(ctx, primitive.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *MongoDB) FindUserByEmail(ctx context.Context, email string) (*models.UserWithCompanyAsObject, error) {
	var user models.UserWithCompanyAsObject
	err := db.userCollection.FindOne(ctx, primitive.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *MongoDB) FindAllUsers(ctx context.Context, companyId primitive.ObjectID) ([]models.UserWithCompanyAsObject, error) {
	var users []models.UserWithCompanyAsObject
	cursor, err := db.userCollection.Find(ctx, primitive.M{"company": companyId})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil

}
