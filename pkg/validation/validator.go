package validation

import "regexp"

var (
	// EmailRX is a regex for sanity checking the format of email addresses.
	// The regex pattern used is taken from  https://html.spec.whatwg.org/#valid-e-mail-address.
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Validator is a simple struct for collecting validation errors.
type Validator struct {
	Errors []ValidationError `json:"errors"`
}

// ValidationError is a simple struct for representing a validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error returns the validation error message.
func (v ValidationError) Error() string {
	return v.Message
}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Check adds an error to the Validator if the condition is false.
func (v *Validator) Check(condition bool, field, message string) {
	if !condition {
		v.Errors = append(v.Errors, ValidationError{Field: field, Message: message})
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
func (v *Validator) In(value string, list []string, key, message string) {
	for i := range list {
		if value == list[i] {
			return
		}
	}
	v.AddError(key, message)
}

// AddError adds an error to the Validator.
func (v *Validator) AddError(field, message string) {
	newError := ValidationError{
		Field:   field,
		Message: message,
	}
	v.Errors = append(v.Errors, newError)
}

// HasErrors returns true if the Validator has any errors.
func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}
