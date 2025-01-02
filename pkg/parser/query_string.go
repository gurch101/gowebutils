package parser

import (
	"math"
	"net/url"
	"strconv"
	"strings"

	"gurch101.github.io/go-web/pkg/validation"
)

// Filters contains common query string parameters for pagination and sort
type Filters struct {
	Page     int
	PageSize int
	Sort     string
}

// PaginationMetadata contains metadata about the current page of paginated data
type PaginationMetadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// ParsePaginationMetadata calculates the pagination metadata based on the total number of records, the current page, and the page size
func ParsePaginationMetadata(totalRecords, page, pageSize int) PaginationMetadata {
	if totalRecords == 0 {
		return PaginationMetadata{}
	}

	return PaginationMetadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

// ParseFilters parses the query string parameters and populates the Filters struct
func (f *Filters) ParseFilters(qs url.Values, v *validation.Validator, sortSafeList []string) (err error) {
	var defaultPage = 1
	var defaultPageSize = 25
	var defaultSort string = "id"
	var page, pageSize *int
	var sort *string
	page, err = ParseInt(qs, "page", &defaultPage)
	if err != nil {
		return err
	}
	f.Page = *page
	pageSize, err = ParseInt(qs, "pageSize", &defaultPageSize)
	if err != nil {
		return err
	}
	f.PageSize = *pageSize
	sort = ParseString(qs, "sort", &defaultSort)
	f.Sort = *sort
	f.validate(v, sortSafeList)
	return nil
}

// Validate checks that the page and page_size parameters contain sensible values and that the sort parameter matches a value in the safelist
func (f *Filters) validate(v *validation.Validator, sortSafeList []string) {
	// Check that the page and page_size parameters contain sensible values.
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	// Check that the sort parameter matches a value in the safelist.
	v.In(f.Sort, sortSafeList, "sort", "invalid sort value")
}

// ParseString returns a string value from the query string or the provided default value if no matching key can be found
func ParseString(qs url.Values, key string, defaultValue *string) *string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	s = strings.TrimSpace(s)
	return &s
}

// ParseInt returns an integer value from the query string or the provided default value if no matching key can be found
func ParseInt(qs url.Values, key string, defaultValue *int) (*int, error) {
	s := qs.Get(key)

	if s == "" {
		return defaultValue, nil
	}

	i, err := strconv.Atoi(s)

	if err != nil {
		return defaultValue, err
	}

	return &i, nil
}

func ParseBool(qs url.Values, key string, defaultValue *bool) *bool {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	result := strings.ToLower(s) == "true"
	return &result
}
