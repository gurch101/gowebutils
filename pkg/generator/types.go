package generator

import (
	"strings"

	"github.com/gurch101/gowebutils/pkg/collectionutils"
	"github.com/gurch101/gowebutils/pkg/stringutils"
)

type Table struct {
	Name          string
	Fields        []Field
	UniqueIndexes []UniqueIndex
	ForeignKeys   []ForeignKey
}

func (t Table) HasUpdateAt() bool {
	return collectionutils.Contains(t.Fields, func(field Field) bool {
		return field.Name == "updated_at"
	})
}

type Field struct {
	Name        string
	DataType    SQLDataType
	Constraints []string
}

type UniqueIndex struct {
	Name   string
	Fields []string
}

type CheckConstraint struct {
	Name       string
	Expression string
}

type ForeignKey struct {
	Table      string
	FromColumn string
	ToColumn   string
	OnDelete   string
	OnUpdate   string
}

func (fk ForeignKey) HumanTableName() string {
	return stringutils.SnakeToHuman(strings.TrimSuffix(fk.Table, "s"))
}

func (fk ForeignKey) SingularTitleCaseTableName() string {
	return stringutils.SnakeToTitle(strings.TrimSuffix(fk.Table, "s"))
}

func (fk ForeignKey) SingularCamelCaseTableName() string {
	return stringutils.SnakeToCamel(strings.TrimSuffix(fk.Table, "s"))
}

func (fk ForeignKey) TitleCaseFromColumnName() string {
	return stringutils.SnakeToTitle(fk.FromColumn)
}

func (fk ForeignKey) JSONName() string {
	return stringutils.SnakeToCamel(fk.FromColumn)
}

type SQLDataType string

const (
	SQLInt           SQLDataType = "Int"
	SQLInt64         SQLDataType = "Int64"
	SQLBoolean       SQLDataType = "Boolean"
	SQLReal          SQLDataType = "Real"
	SQLDecimal       SQLDataType = "Decimal"
	SQLString        SQLDataType = "String"
	SQLDatetime      SQLDataType = "Datetime"
	SQLTimestamp     SQLDataType = "Timestamp"
	SQLDuration      SQLDataType = "Time"
	SQLJson          SQLDataType = "Json"
	SQLVectorFloat32 SQLDataType = "VectorFloat32"
)

//nolint:cyclop
func (s SQLDataType) GoType() string {
	switch s {
	case SQLInt:
		return "int"
	case SQLInt64:
		return "int64"
	case SQLBoolean:
		return "bool"
	case SQLReal:
		return "float64"
	case SQLDecimal:
		return "decimal.Decimal"
	case SQLString:
		return "string"
	case SQLDatetime:
		return "time.Time"
	case SQLTimestamp:
		return "time.Time"
	case SQLDuration:
		return "time.Duration"
	case SQLJson:
		return "json.RawMessage"
	case SQLVectorFloat32:
		return "[]float32"
	default:
		return "unknown"
	}
}

type createHandlerTemplateData struct {
	PackageName           string
	Name                  string
	ModuleName            string
	HumanName             string
	TitleCaseTableName    string
	SingularTitleCaseName string
	SingularCamelCaseName string
	KebabCaseTableName    string
	UniqueConstraint      bool
	IncludeTime           bool
	RequireValidation     bool
	UniqueFields          []UniqueField
	Fields                []RequestField
	ModelFields           []ModelField
	ForeignKeys           []ForeignKey
}

type deleteHandlerTemplateData struct {
	PackageName           string
	Name                  string
	ModuleName            string
	HumanName             string
	TitleCaseTableName    string
	KebabCaseTableName    string
	SingularTitleCaseName string
	SingularCamelCaseName string
	CreateFields          []RequestField
}

type getHandlerTemplateData struct {
	PackageName           string
	Name                  string
	ModuleName            string
	HumanName             string
	SingularTitleCaseName string
	SingularCamelCaseName string
	KebabCaseTableName    string
	ModelFields           []ModelField
	CreateFields          []RequestField
	HasCreatedAt          bool
	HasUpdatedAt          bool
}

type updateHandlerTemplateData struct {
	PackageName           string
	Name                  string
	ModuleName            string
	HumanName             string
	KebabCaseTableName    string
	SingularTitleCaseName string
	SingularCamelCaseName string
	RequireValidation     bool
	ModelFields           []ModelField
	Fields                []RequestField
	ForeignKeys           []ForeignKey
}

type searchHandlerTemplateData struct {
	PackageName           string
	Name                  string
	ModuleName            string
	HumanName             string
	TitleCaseTableName    string
	SingularTitleCaseName string
	SingularCamelCaseName string
	KebabCaseTableName    string
	Fields                []RequestField
	ModelFields           []ModelField
}

type modelTemplateData struct {
	PackageName           string
	Name                  string
	ModuleName            string
	TitleCaseTableName    string
	SingularTitleCaseName string
	SingularCamelCaseName string
	IncludeTime           bool
	IncludeValidation     bool
	ModelFields           []ModelField
	Fields                []RequestField
	UniqueFields          []UniqueField
	ForeignKeys           []ForeignKey
}

type testHelperTemplateData struct {
	PackageName           string
	ModuleName            string
	Name                  string
	TitleCaseTableName    string
	SingularTitleCaseName string
	SingularCamelCaseName string
	HasUpdate             bool
	Fields                []RequestField
	UniqueFields          []UniqueField
	ForeignKeys           []ForeignKey
}

type UniqueField struct {
	Name          string
	TitleCaseName string
	JSONName      string
	HumanName     string
}

func newUniqueField(field Field) UniqueField {
	return UniqueField{
		Name:          field.Name,
		JSONName:      stringutils.SnakeToCamel(field.Name),
		TitleCaseName: stringutils.SnakeToTitle(field.Name),
		HumanName:     stringutils.SnakeToHuman(field.Name),
	}
}

type RequestField struct {
	Name          string
	TitleCaseName string
	JSONName      string
	HumanName     string
	GoType        string
	Required      bool
	IsEmail       bool
}

func IsRequestField(field Field) bool {
	return field.Name != "id" && field.Name != "version" && field.Name != "created_at" && field.Name != "updated_at"
}

func (field RequestField) SwaggerTag() string {
	if field.Required {
		return ` validate:"required"`
	}

	return ""
}

type ModelField struct {
	Name          string
	TitleCaseName string
	CamelCaseName string
	GoType        string
	JSONName      string
}

func newModelField(sanitizedName string, field Field) ModelField {
	return ModelField{
		Name:          field.Name,
		TitleCaseName: stringutils.SnakeToTitle(sanitizedName),
		CamelCaseName: stringutils.SnakeToCamel(sanitizedName),
		GoType:        field.DataType.GoType(),
		JSONName:      stringutils.SnakeToCamel(sanitizedName),
	}
}
