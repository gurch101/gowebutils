package generator

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/fsutils"
)

func ParseSchema(db *dbutils.DBPool) ([]Table, error) {
	tables, err := getDatabaseSchema(db)
	if err != nil {
		return nil, fmt.Errorf("error getting database schema: %w", err)
	}

	return tables, nil
}

func getDatabaseSchema(db *dbutils.DBPool) ([]Table, error) {
	tableNames, err := getTableNames(db)
	if err != nil {
		return nil, err
	}

	var tables []Table

	for _, tableName := range tableNames {
		table, err := processTable(db, tableName)
		if err != nil {
			return nil, err
		}

		tables = append(tables, *table)
	}

	return tables, nil
}

func processTable(db *dbutils.DBPool, tableName string) (*Table, error) {
	tableInfo, err := getTableInfo(db, tableName)
	if err != nil {
		return nil, err
	}

	if err := processUniqueIndexes(db, tableName, tableInfo); err != nil {
		return nil, err
	}

	if err := processCheckConstraints(db, tableName, tableInfo); err != nil {
		return nil, err
	}

	return tableInfo, nil
}

func processUniqueIndexes(db *dbutils.DBPool, tableName string, tableInfo *Table) error {
	uniqueIndexes, err := getUniqueIndexes(db, tableName)
	if err != nil {
		return err
	}

	var tableUniqueIndexes []UniqueIndex

	for _, index := range uniqueIndexes {
		if len(index.Fields) == 1 {
			processSingleFieldUniqueIndex(index, tableInfo)
		} else {
			tableUniqueIndexes = append(tableUniqueIndexes, index)
		}
	}

	tableInfo.UniqueIndexes = tableUniqueIndexes

	return nil
}

func processSingleFieldUniqueIndex(index UniqueIndex, tableInfo *Table) {
	for i, field := range tableInfo.Fields {
		if field.Name == index.Fields[0] {
			field.Constraints = append(field.Constraints, "UNIQUE")
			tableInfo.Fields[i] = field

			break
		}
	}
}

func processCheckConstraints(db *dbutils.DBPool, tableName string, tableInfo *Table) error {
	constraints, err := getCheckConstraints(db, tableName)
	if err != nil {
		return err
	}

	for _, constraint := range constraints {
		applyConstraintToFields(constraint, tableInfo)
	}

	return nil
}

func applyConstraintToFields(constraint CheckConstraint, tableInfo *Table) {
	for i, field := range tableInfo.Fields {
		if field.Name == constraint.Name || strings.Contains(constraint.Expression, field.Name) {
			tableInfo.Fields[i].Constraints = append(tableInfo.Fields[i].Constraints, "CHECK "+constraint.Expression)
			break
		}
	}
}

func getTableNames(db *dbutils.DBPool) ([]string, error) {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get list of tables", err)
	}
	defer fsutils.CloseAndPanic(rows)

	var tableNames []string

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("%w: failed to get table name", err)
		}

		if !strings.HasPrefix(tableName, "sqlite_") && !strings.HasPrefix(tableName, "schema_migrations") && tableName != "sessions" {
			tableNames = append(tableNames, tableName)
		}
	}

	return tableNames, nil
}

func getTableInfo(db *dbutils.DBPool, tableName string) (*Table, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, fmt.Errorf("%w: Failed to get table info", err)
	}
	defer fsutils.CloseAndPanic(rows)

	var fields []Field

	for rows.Next() {
		var (
			cid        int
			name       string
			dataType   string
			notNull    int
			dfltValue  sql.NullString
			primaryKey int
		)

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &primaryKey); err != nil {
			return nil, err
		}

		var constraints []string
		if notNull == 1 {
			constraints = append(constraints, "NOT NULL")
		}

		if primaryKey == 1 {
			constraints = append(constraints, "PRIMARY KEY")
		}

		if dfltValue.Valid {
			constraints = append(constraints, "DEFAULT "+dfltValue.String)
		}

		sqlDataType := mapSQLiteType(dataType)

		if (strings.Contains(name, "id") || strings.Contains(name, "version")) && sqlDataType == SQLInt {
			sqlDataType = SQLInt64
		}

		fields = append(fields, Field{
			Name:        name,
			DataType:    sqlDataType,
			Constraints: constraints,
		})
	}

	return &Table{
		Name:   tableName,
		Fields: fields,
	}, nil
}

func getUniqueIndexes(db *dbutils.DBPool, tableName string) ([]UniqueIndex, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA index_list(%s)", tableName))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to query indexes", err)
	}
	defer fsutils.CloseAndPanic(rows)

	var indexes []UniqueIndex

	for rows.Next() {
		var (
			seq     int
			name    string
			unique  int
			origin  string
			partial int
		)

		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, fmt.Errorf("%w: failed to get indexes", err)
		}

		// Only process unique indexes
		if unique != 1 {
			continue
		}

		// Get the columns in this index
		indexInfo, err := db.Query(fmt.Sprintf("PRAGMA index_info(%s)", name))
		if err != nil {
			return nil, fmt.Errorf("%w: failed to get columns in index", err)
		}

		defer fsutils.CloseAndPanic(indexInfo)

		var columns []string

		for indexInfo.Next() {
			var (
				seqno   int
				cid     int
				colName string
			)

			if err := indexInfo.Scan(&seqno, &cid, &colName); err != nil {
				return nil, fmt.Errorf("%w: failed to get columns in index", err)
			}

			columns = append(columns, colName)
		}

		indexes = append(indexes, UniqueIndex{
			Name:   name,
			Fields: columns,
		})
	}

	return indexes, nil
}

func getCheckConstraints(db *dbutils.DBPool, tableName string) ([]CheckConstraint, error) {
	const (
		sqliteMasterQuery = `
			SELECT sql FROM sqlite_master
			WHERE type='table' AND name=?`
		checkKeyword      = "CHECK"
		constraintKeyword = "CONSTRAINT"
		constraintParts   = 2
	)

	row := db.QueryRow(sqliteMasterQuery, tableName)

	var createSQL string
	if err := row.Scan(&createSQL); err != nil {
		return nil, fmt.Errorf("%w: failed to get create table statement", err)
	}

	var checks []CheckConstraint

	parts := strings.Split(createSQL, ",")

	for _, raw := range parts {
		part := strings.TrimSpace(raw)
		if !strings.Contains(strings.ToUpper(part), checkKeyword) {
			continue
		}

		var name, expression string

		up := strings.ToUpper(part)
		if strings.Contains(up, constraintKeyword) {
			// Example: CONSTRAINT name CHECK (expression)
			nameExpr := strings.SplitN(part, constraintKeyword, constraintParts)
			if len(nameExpr) == constraintParts {
				nameAndCheck := strings.SplitN(nameExpr[1], checkKeyword, constraintParts)
				if len(nameAndCheck) == constraintParts {
					name = strings.TrimSpace(nameAndCheck[0])
					expression = strings.TrimSpace(nameAndCheck[1])
				}
			}
		} else {
			// Example: CHECK (expression)
			checkExpr := strings.SplitN(part, checkKeyword, constraintParts)
			if len(checkExpr) == constraintParts {
				expression = strings.TrimSpace(checkExpr[1])
			}
		}

		expression = strings.TrimPrefix(expression, "(")
		expression = strings.TrimSuffix(expression, ")")

		checks = append(checks, CheckConstraint{
			Name:       name,
			Expression: expression,
		})
	}

	return checks, nil
}

// nolint: cyclop
func mapSQLiteType(sqliteType string) SQLDataType {
	sqliteType = strings.ToUpper(sqliteType)

	if isIntegerType(sqliteType) {
		if strings.Contains(sqliteType, "64") {
			return SQLInt64
		}

		return SQLInt
	}

	if isBooleanType(sqliteType) {
		return SQLBoolean
	}

	if isRealType(sqliteType) {
		return SQLReal
	}

	if isDecimalType(sqliteType) {
		return SQLDecimal
	}

	if isStringType(sqliteType) {
		return SQLString
	}

	if isDateTimeType(sqliteType) {
		return mapDateTimeType(sqliteType)
	}

	if strings.Contains(sqliteType, "JSON") {
		return SQLJson
	}

	if strings.Contains(sqliteType, "BLOB") && strings.Contains(sqliteType, "VECTOR") {
		return SQLVectorFloat32
	}

	return SQLString
}

func isIntegerType(sqliteType string) bool {
	return strings.Contains(sqliteType, "INT")
}

func isBooleanType(sqliteType string) bool {
	return strings.Contains(sqliteType, "BOOL")
}

func isRealType(sqliteType string) bool {
	return strings.Contains(sqliteType, "REAL") ||
		strings.Contains(sqliteType, "FLOAT") ||
		strings.Contains(sqliteType, "DOUBLE")
}

func isDecimalType(sqliteType string) bool {
	return strings.Contains(sqliteType, "DECIMAL") ||
		strings.Contains(sqliteType, "NUMERIC")
}

func isStringType(sqliteType string) bool {
	return strings.Contains(sqliteType, "CHAR") ||
		strings.Contains(sqliteType, "TEXT") ||
		strings.Contains(sqliteType, "CLOB")
}

func isDateTimeType(sqliteType string) bool {
	return strings.Contains(sqliteType, "DATE") || strings.Contains(sqliteType, "TIME")
}

func mapDateTimeType(sqliteType string) SQLDataType {
	if strings.Contains(sqliteType, "TIME") && strings.Contains(sqliteType, "STAMP") {
		return SQLTimestamp
	}

	if strings.Contains(sqliteType, "DATE") {
		return SQLDatetime
	}

	return SQLDuration
}
