package user

import "errors"

var ErrLoginAlreadyExists = errors.New("Login already exists")
var ErrLoginDoesNotExist = errors.New("Login does not exist")
