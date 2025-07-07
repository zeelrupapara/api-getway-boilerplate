package utils

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidatorMessage is a function to generate custom message for validator package
func ValidatorMessage(err error) error {
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err
	}

	validatorErrors := err.(validator.ValidationErrors)

	field := validatorErrors[0].Field() // the field that caused the error
	tag := validatorErrors[0].Tag()     // the validation tag that failed
	param := validatorErrors[0].Param() // the parameter to the validation tag (if any)

	// Generate a custom message based on the field and tag
	var customMessage error
	switch tag {
	case "required":
		customMessage = fmt.Errorf("%s is required", field)
	case "email":
		customMessage = fmt.Errorf("%s is not a valid email", field)
	case "min":
		customMessage = fmt.Errorf("%s must be at least %s characters", field, param)
	case "max":
		customMessage = fmt.Errorf("%s must be at most %s characters", field, param)
	case "gte":
		customMessage = fmt.Errorf("%s must be greater than or equal to %s", field, param)
	case "lte":
		customMessage = fmt.Errorf("%s must be less than or equal to %s", field, param)
	default:
		customMessage = fmt.Errorf("field %s is not valid", field)
	}

	return customMessage
}

// Username validation rule
func UsernameValidation(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	// Define regex pattern for username
	pattern := "^[a-zA-Z0-9_]{3,20}$"
	// patte1rn := "^[A-Za-z][A-Za-z0-9_]{7,29}$"

	matched, _ := regexp.MatchString(pattern, username)
	return matched
}
