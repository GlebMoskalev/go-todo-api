package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		*d = Date{v.UTC().Truncate(24 * time.Hour)}
		return nil
	case nil:
		*d = Date{}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Date", v)
	}
}

func (d Date) Value() (driver.Value, error) {
	return d.Time.UTC().Truncate(24 * time.Hour), nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Time.UTC().Format(time.DateOnly))), nil
}

func (d *Date) UnmarshalJSON(bs []byte) error {
	var s string
	err := json.Unmarshal(bs, &s)
	if err != nil {
		return err
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return err
	}
	*d = Date{t.UTC()}
	return nil
}

type Todo struct {
	ID          int      `json:"id" validate:"omitempty"`
	Title       string   `json:"title" validate:"required,min=3"`
	Description string   `json:"description" validate:"required"`
	Tags        []string `json:"tags" validate:"required"`
	DueDate     *Date    `json:"due_date" validate:"required"`
}

func ValidateTodo(todo Todo) []string {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := validate.Struct(todo)
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
