package parser

import (
	"net/url"
	"strconv"
	"strings"
)

// Filters contains common query string parameters for pagination and sort
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

// ParseFilters parses the query string parameters and populates the Filters struct
func (f *Filters) ParseFilters(qs url.Values) (err error) {
	f.Page, err = ParseInt(qs, "page", 1)
	if err != nil {
		return err
	}
	f.PageSize, err = ParseInt(qs, "pageSize", 10)
	if err != nil {
		return err
	}
	f.Sort = ParseString(qs, "sort", "id")
	return nil
}

// Validate checks that the page and page_size parameters contain sensible values and that the sort parameter matches a value in the safelist
func (f *Filters) Validate() {
	// Check that the page and page_size parameters contain sensible values. v.Check(f.Page > 0, "page", "must be greater than zero")
	// v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	// v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	// v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	// Check that the sort parameter matches a value in the safelist.
	// v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

// ParseString returns a string value from the query string or the provided default value if no matching key can be found
func ParseString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return strings.TrimSpace(s)
}

// ParseInt returns an integer value from the query string or the provided default value if no matching key can be found
func ParseInt(qs url.Values, key string, defaultValue int) (int, error) {
	s := qs.Get(key)

	if s == "" {
		return defaultValue, nil
	}

	i, err := strconv.Atoi(s)

	if err != nil {
		return defaultValue, err
	}

	return i, nil
}
