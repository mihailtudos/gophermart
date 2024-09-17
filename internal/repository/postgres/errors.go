package postgres

import "errors"

var (
	ErrNoRowsFound                     = errors.New("resource not found")
	ErrDuplicateLogin                  = errors.New("duplicate login")
	ErrOrderAlreadyExistsSameUser      = errors.New("the order number has already been uploaded by this user")
	ErrOrderAlreadyExistsDifferentUser = errors.New("the order number has already been uploaded by another user")
	ErrOrderAlreadyAccepted            = errors.New("the order number has already been accepted for processing")
)
