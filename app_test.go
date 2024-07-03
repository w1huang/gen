package gen

import (
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

var dbSchema = "campaign"
var username = "data_owner"
var password = "Secret1!"
var hostname = "localhost"
var port = "5432"
var dbName = "am_api_db"
var tableName = "line_item"
var postgresUrlTemplate = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?search_path=SCHEMA&client_encoding=UTF8", username, password, hostname, port, dbName)

var dataType = map[string]func(columnType gorm.ColumnType) (dataType string){
	"varchar": func(columnType gorm.ColumnType) string {
		nullable, ok := columnType.Nullable()
		if nullable && ok {
			return "sql.NullString"
		} else {
			return columnType.ScanType().Name()
		}
	},
	"text": func(columnType gorm.ColumnType) string {
		nullable, ok := columnType.Nullable()
		if nullable && ok {
			return "sql.NullString"
		} else {
			return columnType.ScanType().Name()
		}
	},
	"int4": func(columnType gorm.ColumnType) string {
		nullable, ok := columnType.Nullable()
		if nullable && ok {
			return "sql.NullInt32"
		} else {
			return columnType.ScanType().Name()
		}
	},
	"int8": func(columnType gorm.ColumnType) string {
		nullable, ok := columnType.Nullable()
		if nullable && ok {
			return "sql.NullInt64"
		} else {
			return columnType.ScanType().Name()
		}
	},
	"timestamptz": func(columnType gorm.ColumnType) string {
		nullable, ok := columnType.Nullable()
		if nullable && ok {
			return "sql.NullTime"
		} else {
			return "time.Time"
		}
	},
}

var modelGenOpts = []ModelOpt{FieldType("deleted_dtm", "gorm.DeletedAt"),
	FieldType("created_dtm", "time.Time"),
	FieldType("modified_dtm", "time.Time"),
	FieldGORMTag("created_dtm", func(tag field.GormTag) field.GormTag { return tag.Set("autoCreateTime", "milli").Remove("default") }),
	FieldGORMTag("modified_dtm", func(tag field.GormTag) field.GormTag { return tag.Set("autoCreateTime", "milli").Remove("default") }),
	FieldGORMTagReg(".+", func(tag field.GormTag) field.GormTag {
		return tag.Remove("column").Remove("comment").Remove("not null")
	}),
	RemoveFieldJSONTagReg(".+"),
	FieldCommentReg(".+", ""),
	// https://github.com/go-gorm/gen/issues/312
	// FieldRelate(field.BelongsTo, "CreditCardRef", card,
	// 	&field.RelateConfig{
	// 		GORMTag: "foreignKey:CreditCardRef",
	// 	}),
}

func mkPostgresUrl(template string, schema string) string {
	return strings.ReplaceAll(template, "SCHEMA", schema)
}

func Test(t *testing.T) {
	var postgres_url string
	var gormdb *gorm.DB

	g := NewGenerator(Config{
		importPkgPaths: []string{"\"gorm.io/gen/models\""},
		// OutPath:           "./query", //producing interfaces
		Mode:              WithoutContext | WithDefaultQuery | WithQueryInterface,
		FieldWithIndexTag: true,
	})
	g.WithDataTypeMap(dataType)
	g.WithFileNameStrategy(func(tableName string) (fileName string) { return tableName })
	// g.WithTableNameStrategy(func(tableName string) (targetTableName string) { return db_schema + "." + tableName })
	// gormdb, _ := gorm.Open(postgres.Open(postgres_url), &gorm.Config{
	// 	NamingStrategy: schema.NamingStrategy{
	// 		TablePrefix:   db_schema + ".", // schema name
	// 		SingularTable: false,
	// 	}})

	// dbSchema = "org"
	// g.ModelPkgPath = "./models/" + dbSchema
	// postgres_url = mkPostgresUrl(postgresUrlTemplate, dbSchema)

	// gormdb, _ = gorm.Open(postgres.Open(postgres_url))
	// g.UseDB(gormdb)

	// account := g.GenerateModel("account", modelGenOpts...)
	// account := g.db.Preload("account")

	dbSchema = "org"
	g.ModelPkgPath = "./models/" + dbSchema
	postgres_url = mkPostgresUrl(postgresUrlTemplate, dbSchema)
	gormdb, _ = gorm.Open(postgres.Open(postgres_url))
	g.UseDB(gormdb)
	organization := g.GenerateModel("organization", modelGenOpts...)

	account := g.GenerateModel("account", append(modelGenOpts,
		FieldRelateModel(field.BelongsTo, "Organization", organization,
			&field.RelateConfig{
				//RelateSlice: true,
				// GORMTag: field.GormTag{}.Set("foriegnKey", "AccountId"),
				GORMTag: field.GormTag{"foreignKey": []string{"OrganizationId"}, "references": []string{"organization_id"}},
				// JSONTag: "",
			}),
		// FieldType("Account", "org.Account"),
	)...)
	// g.Execute()

	g.ApplyBasic(organization, account)
}
