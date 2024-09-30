package template

// Model used as a variable because it cannot load template file after packed, params still can pass file
const Model = `
package {{.StructInfo.Package}}

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	{{range .ImportPkgPaths}}{{.}} ` + "\n" + `{{end}}
)

{{if .TableName -}}const TableName{{.ModelStructName}} = "{{.StructInfo.Package}}.{{.TableName}}"{{- end}}

// {{.ModelStructName}} {{.StructComment}}
type {{.ModelStructName}} struct {
    {{range .Fields}}
    {{if .MultilineComment -}}
	/*
{{.ColumnComment}}
    */
	{{end -}}
    {{.Name}} {{.Type}} ` + "{{.Tags}} " +
	"{{if not .MultilineComment}}{{if .ColumnComment}}// {{.ColumnComment}}{{end}}{{end}}" +
	`{{end}}
}

`

// ModelMethod model struct DIY method
const ModelMethod = `

{{if .Doc -}}// {{.DocComment -}}{{end}}
func ({{.GetBaseStructTmpl}}){{.MethodName}}({{.GetParamInTmpl}})({{.GetResultParamInTmpl}}){{.Body}}
`

const ModelTest = `package tests

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/adgear/ad-manager-api/internal/models/org"
	"github.com/adgear/ad-manager-api/internal/shared_services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type {{.ModelStructName}}TestSuite struct {
	suite.Suite
	db             *sql.DB
	mock           sqlmock.Sqlmock
	dbGorm         *gorm.DB
	{{.ModelStructName}} {{.StructInfo.Package}}.{{.ModelStructName}}
}

func Test{{.ModelStructName}}TestSuite(t *testing.T) {
	suite.Run(t, new({{.ModelStructName}}TestSuite))
}

func (t *{{.ModelStructName}}TestSuite) SetupTest() {
	t.db, t.mock, _ = sqlmock.New()
	t.dbGorm, _ = gorm.Open(postgres.New(postgres.Config{
		Conn: t.db,
	}), &gorm.Config{})
	shared_services.Db = t.dbGorm
}

func (t *{{.ModelStructName}}TestSuite) TestTableName() {
	tableName := t.{{.ModelStructName}}.TableName()
	assert.Equal(t.T(), "{{.StructInfo.Package}}.{{.TableName}}", tableName)
}
`
