package entity

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUsernameExists = errors.New("username already exists")
)

var (
	ErrTodoNotFound = errors.New("todo not found")
)
