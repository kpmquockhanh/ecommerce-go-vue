package repositories

import "errors"

var ErrNotFound = errors.New("resource not found")
var ErrInsufficientStock = errors.New("insufficient stock")
var ErrEmptyCart = errors.New("cart is empty")
