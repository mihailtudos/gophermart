package postgres

import "errors"

var (
	ErrNoRowsFound                     = errors.New("resource not found")
	ErrDuplicateEmail                  = errors.New("duplicate email")
	ErrOrderAlreadyExistsSameUser      = errors.New("the order number has already been uploaded by this user")
	ErrOrderAlreadyExistsDifferentUser = errors.New("the order number has already been uploaded by another user")
	ErrOrderAlreadyAccepted            = errors.New("the order number has already been accepted for processing")
)
