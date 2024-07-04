package gen

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gen/campaign"
	"gorm.io/gen/field"
	"gorm.io/gen/internal/model"
	"gorm.io/gen/lookup"
	"gorm.io/gorm"
)

type DbConfig struct {
	dbName   string
	username string
	password string
	hostname string
	port     int32
	dbSchema string
}

func (dbCfg *DbConfig) WithDbSchema(schema string) {
	dbCfg.dbSchema = schema
}

func (dbCfg *DbConfig) GetDbUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%v/%s?search_path=%s&client_encoding=UTF8", dbCfg.username, dbCfg.password, dbCfg.hostname, dbCfg.port, dbCfg.dbName, dbCfg.dbSchema)
}

func (dbCfg *DbConfig) setDbSchema(schema string, g *Generator) *gorm.DB {
	g.ModelPkgPath = "./" + schema
	dbCfg.WithDbSchema(schema)
	gormdb, _ := gorm.Open(postgres.Open(dbCfg.GetDbUrl()))
	g.UseDB(gormdb)

	return gormdb
}

var dataType = map[string]func(columnType gorm.ColumnType) (dataType string){
	"varchar": func(columnType gorm.ColumnType) string {
		return nullTypeResolver(columnType)
	},
	"text": func(columnType gorm.ColumnType) string {
		return nullTypeResolver(columnType)
	},
	"int4": func(columnType gorm.ColumnType) string {
		return nullTypeResolver(columnType)
	},
	"int8": func(columnType gorm.ColumnType) string {
		return nullTypeResolver(columnType)
	},
	"timestamptz": func(columnType gorm.ColumnType) string {
		return nullTypeResolver(columnType)
	},
}

var nullTypeResolver = func(columnType gorm.ColumnType) string {
	nullable, ok := columnType.Nullable()
	databaseTypeName := columnType.DatabaseTypeName()
	var retVal string
	// Handles nullable type
	if nullable && ok {
		switch databaseTypeName {
		case "varchar":
			retVal = "sql.NullString"
		case "text":
			retVal = "sql.NullString"
		case "int4":
			retVal = "sql.NullInt32"
		case "int8":
			retVal = "sql.NullInt64"
		case "timestamptz":
			retVal = "sql.NullTime"
		}
	} else { // Non-nullable types handler
		if databaseTypeName == "timestamptz" { // Special case for timestamptz
			retVal = "time.Time"
		} else {
			retVal = columnType.ScanType().Name()
		}
	}

	return retVal
}

var modelGenOpts = []ModelOpt{
	FieldType("deleted_dtm", "gorm.DeletedAt"),
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

var RemoveFieldJSONTagReg = func(columnNameReg string) model.ModifyFieldOpt {
	reg := regexp.MustCompile(columnNameReg)
	return func(m *model.Field) *model.Field {
		if reg.MatchString(m.ColumnName) {
			m.Tag.Remove(field.TagKeyJson)
		}
		return m
	}
}

var FieldCommentReg = func(columnNameReg string, comment string) model.ModifyFieldOpt {
	reg := regexp.MustCompile(columnNameReg)
	return func(m *model.Field) *model.Field {
		if reg.MatchString(m.ColumnName) {
			m.ColumnComment = comment
			m.MultilineComment = strings.Contains(comment, "\n")
		}
		return m
	}
}

var BuildFieldRelate = func(g *Generator, tableName string, relModel interface{}) model.CreateFieldOpt {
	ns := g.db.NamingStrategy
	tableSchemName := ns.SchemaName(tableName)
	columnDbName := tableName + "_id"
	columnSchemaName := ns.SchemaName(columnDbName)
	return FieldRelateModel(field.BelongsTo, tableSchemName, relModel,
		&field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{columnSchemaName}, "references": []string{columnDbName}},
			JSONTag: "",
		},
	)
}

func TestAppTest(t *testing.T) {
	var dbConfig DbConfig = DbConfig{
		username: "data_owner",
		password: "Secret1!",
		hostname: "localhost",
		port:     5432,
		dbName:   "am_api_db",
	}

	g := NewGenerator(Config{
		Mode:              WithoutContext | WithDefaultQuery | WithQueryInterface,
		FieldWithIndexTag: true,
	})
	g.WithDataTypeMap(dataType)

	// dbConfig.setDbSchema("lookup", g)
	// g.GenerateModel("inventory_source", modelGenOpts...)

	// dbConfig.setDbSchema("lookup", g)
	// g.GenerateModel("campaign_goal", modelGenOpts...)

	// campaignGoal := lookup.CampaignGoal{}
	// dbConfig.setDbSchema("campaign", g)
	// g.GenerateModel("campaign", append(
	// 	modelGenOpts,
	// 	BuildFieldRelate(g, "campaign_goal", campaignGoal),
	// )...)

	// dbConfig.setDbSchema("campaign", g)
	// g.GenerateModel("targeting", modelGenOpts...)

	inventorySource := lookup.InventorySource{}
	targeting := campaign.Targeting{}
	campaign := campaign.Campaign{}
	dbConfig.setDbSchema("campaign", g)
	g.GenerateModel("line_item", append(
		modelGenOpts,
		BuildFieldRelate(g, "campaign", campaign),
		BuildFieldRelate(g, "targeting", targeting),
		BuildFieldRelate(g, "inventory_source", inventorySource),
	)...)

	g.Execute()
}
