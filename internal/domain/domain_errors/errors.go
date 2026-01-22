package domainerrors

import "errors"

var ErrLoginAlreadyExists = errors.New("login already exists")
var ErrLoginDoesNotExist = errors.New("login does not exist")
var ErrLoginWrongPassword = errors.New("login wrong password")
var ErrOrderIDWrongFormat = errors.New("order id has a wrong format")
var ErrOrderIDExists = errors.New("order id already registred")
var ErrOrderIDBelongsOther = errors.New("order id registred by another user")
