package generator

type Table struct {
	Name          string
	Fields        []Field
	UniqueIndexes []UniqueIndex
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
	SingularTitleCaseName string
	SingularCamelCaseName string
	KebabCaseTableName    string
	ModelFields           []ModelField
	CreateFields          []RequestField
}

type updateHandlerTemplateData struct {
	PackageName           string
	Name                  string
	ModuleName            string
	KebabCaseTableName    string
	SingularTitleCaseName string
	SingularCamelCaseName string
	ModelFields           []ModelField
	Fields                []RequestField
}

type searchHandlerTemplateData struct {
	PackageName           string
	Name                  string
	ModuleName            string
	TitleCaseTableName    string
	SingularTitleCaseName string
	SingularCamelCaseName string
	KebabCaseTableName    string
	Fields                []RequestField
	ModelFields           []ModelField
}

type UniqueField struct {
	Name          string
	TitleCaseName string
	JSONName      string
	HumanName     string
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

type ModelField struct {
	Name          string
	TitleCaseName string
	CamelCaseName string
	GoType        string
	JSONName      string
}
