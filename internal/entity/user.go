package entity

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
}

type UserLogin struct {
	Username string `json:"username" example:"john_doe" validate:"required,min=3,max=20,alphanumunderscore"`
	Password string `json:"password" example:"password123" validate:"required,min=8,passwordstrength"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func alphanumUnderscoreValidation(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return re.MatchString(fl.Field().String())
}

func passwordStrengthValidation(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	re := regexp.MustCompile(`^[A-Za-z\d]{8,}$`)
	if !re.MatchString(password) {
		return false
	}

	hasLetter := false
	hasDigit := false
	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}

func (u *UserLogin) Validate() []string {
	validate := validator.New()
	err := validate.RegisterValidation("alphanumunderscore", alphanumUnderscoreValidation)
	if err != nil {
		return nil
	}
	err = validate.RegisterValidation("passwordstrength", passwordStrengthValidation)
	if err != nil {
		return nil
	}

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err = validate.Struct(u)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			var errList []string
			for _, err := range validationErrors {
				var msg string
				switch err.Tag() {
				case "required":
					msg = fmt.Sprintf("Field '%s' is required", err.Field())
				case "min":
					msg = fmt.Sprintf("Field '%s' must be at least %s character", err.Field(), err.Param())
				case "max":
					msg = fmt.Sprintf("Field '%s' must not exceed %s character", err.Field(), err.Param())
				case "alphanumunderscore":
					msg = fmt.Sprintf("Field '%s' must contain only letters, numbers, or underscores", err.Field())
				case "passwordstrength":
					msg = fmt.Sprintf("Field '%s' must contain at least one letter and one digit", err.Field())
				default:
					msg = fmt.Sprintf("Field '%s' failed validation on %s", err.Field(), err.Param())
				}
				errList = append(errList, msg)
			}
			return errList
		}
	}
	if u.Password == u.Username {
		return []string{"Password must not match username"}
	}
	return nil
}
