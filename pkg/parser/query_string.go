package parser

import (
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/gurch101/gowebutils/pkg/validation"
)

// Filters contains common query string parameters for fieldsets, pagination, and sort.
type Filters struct {
	Fields   []string
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
	fieldsKey   = "fields"
	sortKey     = "sort"
	pageKey     = "page"
	pageSizeKey = "pageSize"
)

// ParseQSMetadata parses the query string sort, page, and page size and populates the Filters struct.
func (f *Filters) ParseQSMetadata(
	queryValues url.Values,
	v *validation.Validator,
	fieldSafeList, sortSafeList []string,
) {
	defaultPage := 1

	defaultPageSize := 25

	defaultSort := "id"

	var sort *string

	page := ParseQSInt(queryValues, pageKey, &defaultPage)

	f.Page = *page

	pageSize := ParseQSInt(queryValues, pageSizeKey, &defaultPageSize)

	f.PageSize = *pageSize

	sort = ParseQSString(queryValues, sortKey, &defaultSort)
	f.Sort = *sort

	f.Fields = ParseQSStringSlice(queryValues, "fields", fieldSafeList)

	f.validate(v, fieldSafeList, sortSafeList)
}

// Validate checks that the page and page_size parameters contain sensible values and
// that the sort parameter matches a value in the safelist.
func (f *Filters) validate(v *validation.Validator, fieldSafeList, sortSafeList []string) {
	const (
		maxPageNumber = 10_000_000
		maxPageSize   = 100
	)
	// Check that the page and page_size parameters contain sensible values.
	v.Check(f.Page > 0, pageKey, "must be greater than zero")
	v.Check(f.Page <= maxPageNumber, pageKey, "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, pageSizeKey, "must be greater than zero")
	v.Check(f.PageSize <= maxPageSize, pageSizeKey, "must be a maximum of 100")
	v.ContainsAll(f.Fields, fieldSafeList, fieldsKey, "invalid field")
	v.In(f.Sort, sortSafeList, sortKey, "invalid sort value")
}

// ParseQSString returns a string value from the query string or the provided
// default value if no matching key can be found.
func ParseQSString(queryValues url.Values, key string, defaultValue *string) *string {
	val := queryValues.Get(key)

	if val == "" {
		return defaultValue
	}

	val = strings.TrimSpace(val)

	return &val
}

// ParseQSInt returns an integer value from the query string or the provided
// default value if no matching key can be found.
func ParseQSInt(queryValues url.Values, key string, defaultValue *int) *int {
	val := queryValues.Get(key)

	if val == "" {
		return defaultValue
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}

	return &intVal
}

// ParseQSInt64 returns an integer value from the query string or the provided
// default value if no matching key can be found.
func ParseQSInt64(queryValues url.Values, key string, defaultValue *int64) *int64 {
	val := queryValues.Get(key)

	if val == "" {
		return defaultValue
	}

	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultValue
	}

	return &intVal
}

// ParseQSBool returns a boolean value from the query string or the provided
// default value if no matching key can be found.
func ParseQSBool(queryValues url.Values, key string, defaultValue *bool) *bool {
	val := queryValues.Get(key)

	if val == "" {
		return defaultValue
	}

	result := strings.ToLower(val) == "true"

	return &result
}

func ParseQSStringSlice(queryValues url.Values, key string, defaultSlice []string) []string {
	val := queryValues.Get(key)
	if val == "" {
		return defaultSlice
	}

	val = strings.TrimSpace(val)

	return strings.Split(val, ",")
}
