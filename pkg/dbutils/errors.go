package dbutils

import (
	"fmt"
	"strings"
)

// ConstraintError represents an error related to database constraints.
type ConstraintErrorType string

const (
	// ConstraintErrorTypeNotNull represents a constraint error related to NOT NULL violations.
	ConstraintErrorTypeNotNull ConstraintErrorType = "not null"
	// ConstraintErrorTypeUnique represents a constraint error related to UNIQUE violations.
	ConstraintErrorTypeUnique ConstraintErrorType = "unique"
	// ConstraintErrorForeignKey represents a constraint error related to FOREIGN KEY violations.
	ConstraintErrorForeignKey ConstraintErrorType = "foreign key"
	// ConstraintErrorCheck represents a constraint error related to CHECK violations.
	ConstraintErrorCheck ConstraintErrorType = "check"
)

// ConstraintError represents an error related to database constraints.
type ConstraintError struct {
	// Type represents the type of constraint error.
	Type ConstraintErrorType
	// Details contains additional information about the constraint error such as field names or the check constraint that failed.
	Details []string
}

// ErrRecordNotFound is returned when a query does not return any rows.
var ErrRecordNotFound = fmt.Errorf("no rows found")

// ErrEditConflict is returned if there is a data race and a conflicting edit made by another user.
var ErrEditConflict = fmt.Errorf("edit conflict")

// Error returns the error message.
func (e ConstraintError) Error() string {
	errMsg := fmt.Sprintf("constraint error: %s %v", e.Type, e.Details)
	return errMsg
}

func (e ConstraintError) DetailContains(d string) bool {
	for _, detail := range e.Details {
		if detail == d {
			return true
		}
	}
	return false
}

func parseError(err error) (*ConstraintError, error) {
	const notNullPrefix = "NOT NULL constraint failed: "
	const uniquePrefix = "UNIQUE constraint failed: "
	const foreignKeyPrefix = "FOREIGN KEY constraint failed"
	const checkPrefix = "CHECK constraint failed: "
	const noRowsPrefix = "sql: no rows in result set"

	input := err.Error()
	var errorType ConstraintErrorType
	var details string

	switch {
	case strings.HasPrefix(input, notNullPrefix):
		errorType = ConstraintErrorTypeNotNull
		details = strings.TrimPrefix(input, notNullPrefix)
	case strings.HasPrefix(input, uniquePrefix):
		errorType = ConstraintErrorTypeUnique
		details = strings.TrimPrefix(input, uniquePrefix)
	case strings.HasPrefix(input, foreignKeyPrefix):
		errorType = ConstraintErrorForeignKey
		details = ""
	case strings.HasPrefix(input, checkPrefix):
		return &ConstraintError{
			Details: []string{strings.TrimPrefix(input, checkPrefix)},
			Type:    ConstraintErrorCheck,
		}, nil
	case strings.HasPrefix(input, noRowsPrefix):
		return nil, ErrRecordNotFound
	default:
		return nil, fmt.Errorf("unhandled error: %s", input)
	}

	var fields []string
	if details != "" {
		// Split the fields by commas and trim whitespace
		parts := strings.Split(details, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		// Extract only the field names (after the table names)
		for _, part := range parts {
			fieldParts := strings.Split(part, ".")
			if len(fieldParts) != 2 {
				return nil, fmt.Errorf("invalid field format: %s", part)
			}
			fields = append(fields, fieldParts[1])
		}
	}

	return &ConstraintError{
		Details: fields,
		Type:    errorType,
	}, nil
}

// WrapDBError returns a ConstraintError if the provided error is a database constraint error.
func WrapDBError(err error) error {
	constraintError, err := parseError(err)
	if err != nil {
		return err
	}
	return *constraintError
}
