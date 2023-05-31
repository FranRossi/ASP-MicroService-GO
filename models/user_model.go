package models

type User struct {
	Name     string `json:"name,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"`
	Email    string `json:"email,omitempty" validate:"required"`
	Role     string `json:"role,omitempty" validate:"required"`
	Company  string `json:"company,omitempty" validate:"required"`
}
