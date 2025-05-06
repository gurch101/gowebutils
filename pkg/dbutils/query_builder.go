package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gurch101/gowebutils/pkg/fsutils"
	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const execTimeout = 3 * time.Second

type QueryBuilder struct {
	selectFields []string
	table        string
	joins        []string
	conditions   []string
	args         []interface{}
	groupBy      []string
	orderBy      []string
	limit        int
	offset       int
	db           DB
}

type QueryOperator string

const (
	OpStartsWith QueryOperator = "starts_with"
	OpContains   QueryOperator = "contains"
	OpEndsWith   QueryOperator = "ends_with"
)

// NewQueryBuilder creates a new QueryBuilder instance which can be used to build and execute SQL queries.
func NewQueryBuilder(db DB) *QueryBuilder {
	return &QueryBuilder{
		selectFields: []string{},
		table:        "",
		joins:        []string{},
		conditions:   []string{},
		args:         []interface{}{},
		groupBy:      []string{},
		orderBy:      []string{},
		limit:        -1, // Default to no limit
		offset:       -1, // Default to no offset
		db:           db,
	}
}

// Select sets the fields to be selected in the query.
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.selectFields = append(qb.selectFields, fields...)

	return qb
}

// From sets the table to be queried.
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.table = table

	return qb
}

// Join adds a JOIN clause to the query.
func (qb *QueryBuilder) Join(joinType, table, onCondition string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("%s JOIN %s ON %s", joinType, table, onCondition))

	return qb
}

// Where adds a WHERE clause to the query.
// condition should be in the format of "field = ?" or "field IN (?, ?, ?)"
// and args should be the values to be bound to the condition.
func (qb *QueryBuilder) Where(condition string, args ...any) *QueryBuilder {
	if len(args) > 0 && !isNilValue(args[0]) {
		qb.conditions = append(qb.conditions, parenthesize(condition))
		qb.args = append(qb.args, args[0])
	}

	return qb
}

// WhereLike adds a WHERE clause to the query with a LIKE pattern.
func (qb *QueryBuilder) WhereLike(condition string, patternType QueryOperator, value *string) *QueryBuilder {
	if value == nil {
		return qb
	}

	pattern := generateLikePattern(patternType, *value)

	qb.addLikeCondition(condition, "")
	qb.args = append(qb.args, pattern)

	return qb
}

// AndWhere adds a WHERE clause to the query with an AND conjunction.
func (qb *QueryBuilder) AndWhere(condition string, args ...interface{}) *QueryBuilder {
	if len(args) > 0 && !isNilValue(args[0]) {
		qb.addCondition(condition, "AND")
		qb.args = append(qb.args, args...)
	}

	return qb
}

// AndWhereLike adds a WHERE clause to the query with an AND conjunction and a LIKE pattern.
func (qb *QueryBuilder) AndWhereLike(condition string, patternType QueryOperator, value *string) *QueryBuilder {
	if value == nil {
		return qb
	}

	pattern := generateLikePattern(patternType, *value)

	qb.addLikeCondition(condition, "AND")
	qb.args = append(qb.args, pattern)

	return qb
}

// OrWhere adds a WHERE clause to the query with an OR conjunction.
func (qb *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder {
	if len(args) > 0 && !isNilValue(args[0]) {
		qb.addCondition(condition, "OR")
		qb.args = append(qb.args, args...)
	}

	return qb
}

// OrWhereLike adds a WHERE clause to the query with an OR conjunction and a LIKE pattern.
func (qb *QueryBuilder) OrWhereLike(condition string, patternType QueryOperator, value *string) *QueryBuilder {
	if value == nil {
		return qb
	}

	pattern := generateLikePattern(patternType, *value)

	qb.addLikeCondition(condition, "OR")
	qb.args = append(qb.args, pattern)

	return qb
}

// GroupBy adds a GROUP BY clause to the query.
func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, fields...)

	return qb
}

// fields will be -<name> for descending order and <name> for ascending order.
func (qb *QueryBuilder) OrderBy(fields ...string) *QueryBuilder {
	// fields will be -<name> for descending order and <name> for ascending order
	// e.g. "name", "-age"
	for i, field := range fields {
		if strings.HasPrefix(field, "-") {
			fields[i] = stringutils.CamelToSnake(strings.TrimPrefix(field, "-")) + " DESC"
		} else {
			fields[i] = stringutils.CamelToSnake(field) + " ASC"
		}
	}

	qb.orderBy = append(qb.orderBy, fields...)

	return qb
}

// Limit sets the maximum number of rows to return.
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit

	return qb
}

// Offset sets the number of rows to skip before returning results.
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset

	return qb
}

// Page sets the page number and page size for pagination.
func (qb *QueryBuilder) Page(page, pageSize int) *QueryBuilder {
	qb.offset = (page - 1) * pageSize
	qb.limit = pageSize

	return qb
}

// Build generates the SQL query and returns it along with the arguments.
func (qb *QueryBuilder) Build() (string, []interface{}) {
	if qb.table == "" {
		panic("Table not specified")
	}

	query := strings.Builder{}

	// SELECT clause
	if len(qb.selectFields) > 0 {
		query.WriteString("SELECT ")
		query.WriteString(strings.Join(qb.selectFields, ", "))
	} else {
		query.WriteString("SELECT *")
	}

	// FROM clause
	query.WriteString(" FROM ")
	query.WriteString(qb.table)

	// JOIN clauses
	if len(qb.joins) > 0 {
		query.WriteString(" ")
		query.WriteString(strings.Join(qb.joins, " "))
	}

	// WHERE clause
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.conditions, " "))
	}

	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// ORDER BY clause
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderBy, ", "))
	}

	// LIMIT and OFFSET clauses
	if qb.limit >= 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	if qb.offset >= 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), qb.args
}

// Query executes the query and calls the callback function for each row.
func (qb *QueryBuilder) Query(callback func(*sql.Rows) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	return qb.QueryContext(ctx, callback)
}

// QueryContext executes the query with the given context and calls the callback function for each row.
func (qb *QueryBuilder) QueryContext(ctx context.Context, callback func(*sql.Rows) error) error {
	query, args := qb.Build()
	rows, err := qb.db.QueryContext(ctx, query, args...)

	if err != nil {
		return fmt.Errorf("query builder exec error: %w", err)
	}

	defer fsutils.CloseAndPanic(rows)

	for rows.Next() {
		if err := callback(rows); err != nil {
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("query builder rows error: %w", err)
	}

	return nil
}

// Query executes the query and binds the results to the provided destination.
func (qb *QueryBuilder) QueryRow(dest ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	return qb.QueryRowContext(ctx, dest...)
}

// QueryRowContext executes the query with the given context and binds the results to the provided destination.
func (qb *QueryBuilder) QueryRowContext(ctx context.Context, dest ...any) error {
	query, args := qb.Build()

	err := qb.db.QueryRowContext(ctx, query, args...).Scan(dest...)
	if err != nil {
		return WrapDBError(err)
	}

	return nil
}

// Helper function to check if the value inside an interface{} is nil.
func isNilValue(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	// Check if it's a pointer, channel, func, interface, map, or slice
	if val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		return val.IsNil()
	}

	return false
}

func parenthesize(s string) string {
	return fmt.Sprintf("(%s)", s)
}

func generateLikePattern(patternType QueryOperator, value string) string {
	switch patternType {
	case OpStartsWith:
		return value + "%" // e.g., "abc%"
	case OpEndsWith:
		return "%" + value // e.g., "%abc"
	case OpContains:
		return fmt.Sprintf("%%%s%%", value) // e.g., "%abc%"
	default:
		panic("Invalid pattern type: use 'starts_with', 'ends_with', or 'contains'")
	}
}

func (qb *QueryBuilder) addLikeCondition(condition, conjunction string) {
	qb.addCondition(condition+" LIKE ?", conjunction)
}

func (qb *QueryBuilder) addCondition(condition, conjunction string) {
	if len(qb.conditions) > 0 {
		lastConditionIndex := len(qb.conditions) - 1
		formattedCondition := fmt.Sprintf("%s %s (%s)", qb.conditions[lastConditionIndex], conjunction, condition)
		qb.conditions[lastConditionIndex] = formattedCondition
	} else {
		qb.conditions = append(qb.conditions, parenthesize(condition))
	}
}
