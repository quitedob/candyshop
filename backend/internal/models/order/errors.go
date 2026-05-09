package order

import "errors"

// ErrInsufficientStock is returned when stock reservation or dispatch fails
// because available quantity is insufficient.
var ErrInsufficientStock = errors.New("insufficient stock")
