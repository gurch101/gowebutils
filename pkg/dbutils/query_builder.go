package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gurch101.github.io/go-web/pkg/stringutils"
)

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
	db           *sql.DB
}

type QueryOperator string

const (
	OpStartsWith QueryOperator = "starts_with"
	OpContains   QueryOperator = "contains"
	OpEndsWith   QueryOperator = "ends_with"
)

func NewQueryBuilder(db *sql.DB) *QueryBuilder {
	return &QueryBuilder{
		selectFields: []string{},
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

func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.selectFields = append(qb.selectFields, fields...)
	return qb
}

func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.table = table
	return qb
}

func (qb *QueryBuilder) Join(joinType, table, onCondition string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("%s JOIN %s ON %s", joinType, table, onCondition))
	return qb
}

func (qb *QueryBuilder) Where(condition string, args ...any) *QueryBuilder {
	if len(args) > 0 && !isNilValue(args[0]) {
		qb.conditions = append(qb.conditions, fmt.Sprintf("(%s)", condition))
		qb.args = append(qb.args, args[0])
	}
	return qb
}

func generateLikePattern(patternType QueryOperator, value string) string {
	switch patternType {
	case "starts_with":
		return fmt.Sprintf("%s%%", value) // e.g., "abc%"
	case "ends_with":
		return fmt.Sprintf("%%%s", value) // e.g., "%abc"
	case "contains":
		return fmt.Sprintf("%%%s%%", value) // e.g., "%abc%"
	default:
		panic("Invalid pattern type: use 'starts_with', 'ends_with', or 'contains'")
	}
}

func (qb *QueryBuilder) WhereLike(condition string, patternType QueryOperator, value *string) *QueryBuilder {
	if value == nil {
		return qb
	}

	pattern := generateLikePattern(patternType, *value)
	qb.conditions = append(qb.conditions, fmt.Sprintf("(%s LIKE ?)", condition))
	qb.args = append(qb.args, pattern)
	return qb
}

func (qb *QueryBuilder) AndWhere(condition string, args ...interface{}) *QueryBuilder {
	if len(args) > 0 && !isNilValue(args[0]) {
		if len(qb.conditions) > 0 {
			qb.conditions[len(qb.conditions)-1] = fmt.Sprintf("%s AND (%s)", qb.conditions[len(qb.conditions)-1], condition)
		} else {
			qb.conditions = append(qb.conditions, fmt.Sprintf("(%s)", condition))
		}
		qb.args = append(qb.args, args...)
	}
	return qb
}

func (qb *QueryBuilder) AndWhereLike(condition string, patternType QueryOperator, value *string) *QueryBuilder {
	if value == nil {
		return qb
	}

	pattern := generateLikePattern(patternType, *value)
	if len(qb.conditions) > 0 {
		qb.conditions[len(qb.conditions)-1] = fmt.Sprintf("%s AND (%s LIKE ?)", qb.conditions[len(qb.conditions)-1], condition)
	} else {
		qb.conditions = append(qb.conditions, fmt.Sprintf("(%s LIKE ?)", condition))
	}
	qb.args = append(qb.args, pattern)
	return qb
}

func (qb *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder {
	if len(args) > 0 && !isNilValue(args[0]) {
		if len(qb.conditions) > 0 {
			qb.conditions[len(qb.conditions)-1] = fmt.Sprintf("%s OR (%s)", qb.conditions[len(qb.conditions)-1], condition)
		} else {
			qb.conditions = append(qb.conditions, fmt.Sprintf("(%s)", condition))
		}
		qb.args = append(qb.args, args...)
	}
	return qb
}

func (qb *QueryBuilder) OrWhereLike(condition string, patternType QueryOperator, value *string) *QueryBuilder {
	if value == nil {
		return qb
	}

	pattern := generateLikePattern(patternType, *value)
	if len(qb.conditions) > 0 {
		qb.conditions[len(qb.conditions)-1] = fmt.Sprintf("%s OR (%s LIKE ?)", qb.conditions[len(qb.conditions)-1], condition)
	} else {
		qb.conditions = append(qb.conditions, fmt.Sprintf("(%s LIKE ?)", condition))
	}
	qb.args = append(qb.args, pattern)
	return qb
}
func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, fields...)
	return qb
}

func (qb *QueryBuilder) OrderBy(fields ...string) *QueryBuilder {
	// fields will be -<name> for descending order and <name> for ascending order
	// e.g. "name", "-age"
	for i, field := range fields {
		if strings.HasPrefix(field, "-") {
			fields[i] = fmt.Sprintf("%s DESC", stringutils.CamelToSnake(strings.TrimPrefix(field, "-")))
		} else {
			fields[i] = fmt.Sprintf("%s ASC", stringutils.CamelToSnake(field))
		}
	}
	qb.orderBy = append(qb.orderBy, fields...)
	return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

func (qb *QueryBuilder) Page(page, pageSize int) *QueryBuilder {
	qb.offset = (page - 1) * pageSize
	qb.limit = pageSize
	return qb
}

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

func (qb *QueryBuilder) Execute(callback func(*sql.Rows) error) error {
	query, args := qb.Build()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := qb.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := callback(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// Helper function to check if the value inside an interface{} is nil
func isNilValue(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case *int:
		return val == nil
	case *string:
		return val == nil
	case *bool:
		return val == nil
	case *float64:
		return val == nil
	case *struct{}:
		return val == nil
	case *interface{}:
		return val == nil
	// Add other pointer types as needed
	default:
		return false
	}
}
