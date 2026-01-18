package user

import "errors"

var ErrLoginAlreadyExists = errors.New("login already exists")
var ErrLoginDoesNotExist = errors.New("login does not exist")
var ErrLoginWrongPassword = errors.New("login wrong password")
