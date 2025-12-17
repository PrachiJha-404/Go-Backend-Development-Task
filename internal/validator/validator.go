package validator

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator with custom logic
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator with custom validation rules
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validation rules
	v.RegisterValidation("dateformat", validateDateFormat)
	v.RegisterValidation("notfuture", validateNotFuture)

	return &Validator{validate: v}
}

// ValidateStruct validates a struct and returns formatted error messages
func (v *Validator) ValidateStruct(data interface{}) error {
	if err := v.validate.Struct(data); err != nil {
		return fmt.Errorf(formatValidationErrors(err))
	}
	return nil
}

// validateDateFormat checks if a date string is in YYYY-MM-DD format and is a valid date
func validateDateFormat(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// validateNotFuture checks if a date is not in the future
func validateNotFuture(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	dob, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	return dob.Before(time.Now())
}

// formatValidationErrors converts validator errors into user-friendly messages
func formatValidationErrors(err error) string {
	var messages []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range validationErrors {
			msg := getErrorMessage(fe)
			messages = append(messages, msg)
		}
	}
	return strings.Join(messages, "; ")
}

// getErrorMessage returns a user-friendly error message for a validation error
func getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "dateformat":
		return fmt.Sprintf("%s must be in YYYY-MM-DD format", field)
	case "notfuture":
		return fmt.Sprintf("%s cannot be in the future", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
