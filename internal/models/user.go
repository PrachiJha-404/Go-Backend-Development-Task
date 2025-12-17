package models

import "time"

type UserResponse struct {
	ID   int32     `json:"id"`
	Name string    `json:"name"`
	DOB  time.Time `json:"dob"`
	Age  int       `json:"age"`
}

// CreateUserRequest is what we expect from the user when they POST
type CreateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	DOB  string `json:"dob" validate:"required,dateformat,notfuture"` // We keep this as string to parse it later
}

// UpdateUserRequest is what we expect when they PUT
type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	DOB  string `json:"dob" validate:"required,dateformat,notfuture"`
}
