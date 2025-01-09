package parser

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/gurch101/gowebutils/pkg/validation"
)

// Filters contains common query string parameters for pagination and sort.
type Filters struct {
	Page     int
	PageSize int
	Sort     string
}

// PaginationMetadata contains metadata about the current page of paginated data.
type PaginationMetadata struct {
	CurrentPage  int `json:"currentPage,omitempty"`
	PageSize     int `json:"pageSize,omitempty"`
	FirstPage    int `json:"firstPage,omitempty"`
	LastPage     int `json:"lastPage,omitempty"`
	TotalRecords int `json:"totalRecords,omitempty"`
}

// ParsePaginationMetadata calculates the pagination metadata based on the total number of records,
// the current page, and the page size.
func ParsePaginationMetadata(totalRecords, page, pageSize int) PaginationMetadata {
	if totalRecords == 0 {
		return PaginationMetadata{
			CurrentPage:  0,
			PageSize:     0,
			FirstPage:    0,
			LastPage:     0,
			TotalRecords: 0,
		}
	}

	return PaginationMetadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

const (
	sortKey     = "sort"
	pageKey     = "page"
	pageSizeKey = "pageSize"
)

// ParseFilters parses the query string parameters and populates the Filters struct.
func (f *Filters) ParseFilters(queryValues url.Values, v *validation.Validator, sortSafeList []string) {
	defaultPage := 1

	defaultPageSize := 25

	defaultSort := "id"

	var page, pageSize *int

	var sort *string

	page, err := ParseInt(queryValues, pageKey, &defaultPage)
	if err != nil {
		v.AddError(pageKey, "Invalid page")

		return
	}

	f.Page = *page

	pageSize, err = ParseInt(queryValues, pageSizeKey, &defaultPageSize)
	if err != nil {
		v.AddError(pageSizeKey, "Invalid pageSize")

		return
	}

	f.PageSize = *pageSize
	sort = ParseString(queryValues, sortKey, &defaultSort)
	f.Sort = *sort
	f.validate(v, sortSafeList)
}

// Validate checks that the page and page_size parameters contain sensible values and
// that the sort parameter matches a value in the safelist.
func (f *Filters) validate(v *validation.Validator, sortSafeList []string) {
	const (
		maxPageNumber = 10_000_000
		maxPageSize   = 100
	)
	// Check that the page and page_size parameters contain sensible values.
	v.Check(f.Page > 0, pageKey, "must be greater than zero")
	v.Check(f.Page <= maxPageNumber, pageKey, "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, pageSizeKey, "must be greater than zero")
	v.Check(f.PageSize <= maxPageSize, pageSizeKey, "must be a maximum of 100")
	// Check that the sort parameter matches a value in the safelist.
	v.In(f.Sort, sortSafeList, sortKey, "invalid sort value")
}

// ParseString returns a string value from the query string or the provided
// default value if no matching key can be found.
func ParseString(queryValues url.Values, key string, defaultValue *string) *string {
	val := queryValues.Get(key)

	if val == "" {
		return defaultValue
	}

	val = strings.TrimSpace(val)

	return &val
}

// ParseInt returns an integer value from the query string or the provided
// default value if no matching key can be found.
func ParseInt(queryValues url.Values, key string, defaultValue *int) (*int, error) {
	val := queryValues.Get(key)

	if val == "" {
		return defaultValue, nil
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue, fmt.Errorf("invalid value for env var %s: %w", key, err)
	}

	return &intVal, nil
}

// ParseBool returns a boolean value from the query string or the provided
// default value if no matching key can be found.
func ParseBool(queryValues url.Values, key string, defaultValue *bool) *bool {
	val := queryValues.Get(key)

	if val == "" {
		return defaultValue
	}

	result := strings.ToLower(val) == "true"

	return &result
}
