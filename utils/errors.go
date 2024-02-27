package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func BuildErrFromValidator(validationErrors validator.ValidationErrors) error {
	if len(validationErrors) == 0 {
		return nil
	}
	preparedErrors := []string{}
	for _, fieldErr := range validationErrors {
		preparedError := fieldErr.Error()
		if fieldErr.Tag() == "required" || fieldErr.Tag() == "gt" {
			preparedError = fmt.Sprintf("setup %s field", fieldErr.Field())
		}
		preparedErrors = append(preparedErrors, preparedError)
	}

	//nolint: goerr113
	return fmt.Errorf(strings.Join(preparedErrors, "; "))
}

func JSONFieldNameForValidator(fld reflect.StructField) string {
	jsonTagWithAnnotationLen := 2
	name := strings.SplitN(fld.Tag.Get("json"), ",", jsonTagWithAnnotationLen)[0]
	if name == "-" {
		return ""
	}

	return name
}
