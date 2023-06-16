package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id         string `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string `json:"name,omitempty" validate:"required"`
	Password   string `json:"password,omitempty" validate:"required"`
	Email      string `json:"email,omitempty" validate:"required"`
	Role       string `json:"role,omitempty" validate:"required"`
	Company    string `json:"company,omitempty" validate:"required"`
	Invitation bool   `json:"invitation,omitempty"`
}

type UserWithCompanyAsObject struct {
	Id       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty"`
	Password string             `json:"password,omitempty"`
	Email    string             `json:"email,omitempty"`
	Role     string             `json:"role,omitempty"`
	Company  primitive.ObjectID `json:"company"`
}
