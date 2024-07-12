package models

import (
	"errors"

	"github.com/pocketbase/pocketbase/models"
)

type User struct {
	models.Model

	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

func NewUser() *User {
	return &User{
		Model: models.Model{
			Collection: "users",
		},
	}
}

func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
