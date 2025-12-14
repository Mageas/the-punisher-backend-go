package utils

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func ValidateStruct(inputStruct any) error {
	return validate.Struct(inputStruct)
}
