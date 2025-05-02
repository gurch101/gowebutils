// Package validation provides a validator for validating user input.
package validation

import (
	"regexp"
	"slices"
)

// EmailRX is a regex for sanity checking the format of email addresses.
// The regex pattern used is taken from  https://html.spec.whatwg.org/#valid-e-mail-address.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$") //nolint:lll

// Validator is a simple struct for collecting validation errors.
type Validator struct {
	Errors []Error `json:"errors"`
}

// Error is a simple struct for representing a validation error.
type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error returns the validation error message.
func (v Error) Error() string {
	return v.Message
}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{
		Errors: []Error{},
	}
}

// Check adds an error to the Validator if the condition is false.
func (v *Validator) Check(condition bool, field, message string) {
	if !condition {
		v.Errors = append(v.Errors, Error{Field: field, Message: message})
	}
}

// ContainsAll checks if the second slice contains all elements of the first slice.
func (v *Validator) ContainsAll(values, list []string, key, message string) {
	for i := range values {
		// add error and return if the value is not in the list
		if !v.In(values[i], list, key, message+": "+values[i]) {
			return
		}
	}
}

// Matches returns true if a string value matches a specific regexp pattern.
func (v *Validator) Matches(value string, rx *regexp.Regexp, field, message string) {
	v.Check(rx.MatchString(value), field, message)
}

func (v *Validator) Email(value string, field, message string) {
	v.Check(EmailRX.MatchString(value), field, message)
}

func (v *Validator) Required(value, field, message string) {
	v.Check(value != "", field, message)
}

// In returns true if a specific value is in a list of strings.
func (v *Validator) In(value string, list []string, key, message string) bool {
	if slices.Contains(list, value) {
		return true
	}

	v.AddError(key, message)

	return false
}

// AddError adds an error to the Validator.
func (v *Validator) AddError(field, message string) {
	newError := Error{
		Field:   field,
		Message: message,
	}
	v.Errors = append(v.Errors, newError)
}

// HasErrors returns true if the Validator has any errors.
func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}
