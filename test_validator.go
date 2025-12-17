package main

import (
	"fmt"
	"user-api/internal/models"
	"user-api/internal/validator"
)

func main() {
	v := validator.NewValidator()

	// Test 1: Valid request
	validReq := models.CreateUserRequest{
		Name: "John Doe",
		DOB:  "1990-01-15",
	}
	if err := v.ValidateStruct(validReq); err != nil {
		fmt.Printf("Test 1 FAILED: %v\n", err)
	} else {
		fmt.Println("Test 1 PASSED: Valid request accepted")
	}

	// Test 2: Empty name
	emptyNameReq := models.CreateUserRequest{
		Name: "",
		DOB:  "1990-01-15",
	}
	if err := v.ValidateStruct(emptyNameReq); err != nil {
		fmt.Printf("Test 2 PASSED: Empty name rejected - %v\n", err)
	} else {
		fmt.Println("Test 2 FAILED: Empty name should be rejected")
	}

	// Test 3: Invalid date format
	invalidDateReq := models.CreateUserRequest{
		Name: "Jane Doe",
		DOB:  "01-15-1990",
	}
	if err := v.ValidateStruct(invalidDateReq); err != nil {
		fmt.Printf("Test 3 PASSED: Invalid date rejected - %v\n", err)
	} else {
		fmt.Println("Test 3 FAILED: Invalid date should be rejected")
	}

	// Test 4: Future date
	futureReq := models.CreateUserRequest{
		Name: "Future Person",
		DOB:  "2099-01-15",
	}
	if err := v.ValidateStruct(futureReq); err != nil {
		fmt.Printf("Test 4 PASSED: Future date rejected - %v\n", err)
	} else {
		fmt.Println("Test 4 FAILED: Future date should be rejected")
	}

	// Test 5: Name too long (over 255 chars)
	longNameReq := models.CreateUserRequest{
		Name: "a" + string(make([]byte, 255)) + "a",
		DOB:  "1990-01-15",
	}
	if err := v.ValidateStruct(longNameReq); err != nil {
		fmt.Printf("Test 5 PASSED: Long name rejected - %v\n", err)
	} else {
		fmt.Println("Test 5 FAILED: Long name should be rejected")
	}
}
