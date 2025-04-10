package apiUtils

import (
	"backend/dto"
	"backend/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func AddCustomValidator(val any) error {
	v, ok := val.(*validator.Validate)
	if !ok {
		return fmt.Errorf("cannot convert validator to go-playground/validator/v10")
	}

	v.RegisterTagNameFunc(validatorTagFunc)

	return nil
}

func CustomValidationError(err validator.FieldError) string {
	fieldName := strings.Join(strings.Split(err.Namespace(), ".")[1:], ".")
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "email":
		return fmt.Sprintf("%s not in email format", fieldName)
	case "min":
		return fmt.Sprintf("%s must be longer than or equal %s characters", fieldName, err.Param())
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s characters", fieldName, err.Param())
	case "gt":
		return fmt.Sprintf("%s should be atleast %s", fieldName, err.Param())
	case "gte":
		return fmt.Sprintf("%s should have atleast %s element", fieldName, err.Param())
	case "lte":
		return fmt.Sprintf("%s should have maximum %s elements", fieldName, err.Param())
	case "iso3166_1_alpha3":
		return fmt.Sprintf("%s is not valid country '%s'", fieldName, err.Param())
	default:
		return err.Error()
	}
}

func validatorTagFunc(fl reflect.StructField) string {
	name := strings.SplitN(fl.Tag.Get("json"), ",", 2)
	if len(name) > 1 && name[1] == "-" {
		return ""
	}
	return name[0]
}

func ValidatorError(err error) error {
	errResponse := dto.Error{}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fieldErr := range ve {
			errResponse.AddReason(CustomValidationError(fieldErr))
		}
	} else {
		errResponse.AddReason(err.Error())
	}
	errResponse.Code = http.StatusBadRequest
	return &errResponse
}

func ValidateStruct(val any) error {
	validate := validator.New()
	validate.SetTagName("binding")

	err := validate.Struct(val)
	if err != nil {
		return ValidatorError(err)
	}
	return nil
}

func AssignAndValidateCreateUserPayload(c context.Context, p any, payload any) error {
	err := utils.ConvertToJSONAndBack(p, &payload)
	if err != nil {
		return dto.NewErrorWithStatus(http.StatusBadRequest, "invalid payload")
	}

	err = ValidateStruct(payload)
	if err != nil {
		slog.ErrorContext(c, "validation error", slog.Any("errors", err))
		return err
	}

	return nil
}
