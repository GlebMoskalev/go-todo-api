package entity

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

type Todo struct {
	ID          int      `json:"id" validate:"omitempty"`
	Title       string   `json:"title" validate:"required,min=3"`
	Description string   `json:"description" validate:"required"`
	Tags        []string `json:"tags" validate:"required"`
	DueDate     *Date    `json:"due_date" validate:"required"`
}

func (t *Todo) Validate() []string {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := validate.Struct(t)
	if err != nil {
		var validationErrors validator.ValidationErrors
		errors.As(err, &validationErrors)
		var errList []string
		for _, err := range validationErrors {
			var msg string
			switch err.Tag() {
			case "required":
				msg = fmt.Sprintf("Field '%s' is required", err.Field())
			case "min":
				msg = fmt.Sprintf("Filed '%s' must be least %s characters", err.Field(), err.Param())
			default:
				msg = fmt.Sprintf("Field %s failled validation on %s", err.Field(), err.Tag())
			}
			errList = append(errList, msg)
		}
		return errList
	}
	return nil
}
