package validator

import (
	"errors"
	"regexp"
	"unicode"
	"unicode/utf8"
)

// TODO - low sev, could be replaced with https://pkg.go.dev/github.com/go-playground/validator/v10
var (
	EmailRX = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z]{2,})+$`)

	ErrValidation = errors.New("validation error")
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}

func IsValidOrderNumber(number string) bool {
	// invalid number
	if utf8.RuneCountInString(number) < 2 {
		return false
	}

	var sum int

	// Use for range to iterate over the string from left to right
	for i, digit := range number {
		if !unicode.IsDigit(digit) {
			return false
		}

		// Get the integer value of the digit
		n := int(digit - '0')

		// Since the Luhn algorithm typically operates from right to left,
		// reverse the alternate pattern using the index.
		if (len(number)-i)%2 == 0 {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}

		sum += n
	}

	return sum%10 == 0
}
