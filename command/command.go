package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
)

var MysqlToGotype = map[string]string{
	"int":                "int64",
	"integer":            "int64",
	"tinyint":            "int64",
	"smallint":           "int64",
	"mediumint":          "int64",
	"bigint":             "int64",
	"int unsigned":       "int64",
	"integer unsigned":   "int64",
	"tinyint unsigned":   "int64",
	"smallint unsigned":  "int64",
	"mediumint unsigned": "int64",
	"bigint unsigned":    "int64",
	"bit":                "int64",
	"bool":               "bool",
	"enum":               "string",
	"set":                "string",
	"varchar":            "string",
	"char":               "string",
	"tinytext":           "string",
	"mediumtext":         "string",
	"text":               "string",
	"longtext":           "string",
	"blob":               "string",
	"tinyblob":           "string",
	"mediumblob":         "string",
	"longblob":           "string",
	"date":               "carbon.Date",
	"datetime":           "carbon.DateTime",
	"timestamp":          "carbon.Timestamp",
	"time":               "string",
	"float":              "float64",
	"double":             "float64",
	"decimal":            "float64",
	"binary":             "string",
	"varbinary":          "string",
	"json":               "json.RawMessage",
}

type databaseInfo struct {
	Tables []schema.Table `gorm:"-"`
}
type ModelCreateCommand struct {
}

func NewCreateModelCommand() *ModelCreateCommand {
	return &ModelCreateCommand{}
}

// Signature The name and signature of the console command.
func (receiver *ModelCreateCommand) Signature() string {
	return "create:model"
}

// Description The console command description.
func (receiver *ModelCreateCommand) Description() string {
	return "Create model"
}

// Extend The console command extend.
func (receiver *ModelCreateCommand) Extend() command.Extend {
	return command.Extend{
		Category: "create",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "table",
				Value:   "",
				Aliases: []string{"t"},
				Usage:   "model table name",
			},
			&command.StringFlag{
				Name:    "database",
				Value:   "",
				Aliases: []string{"d"},
				Usage:   "database name",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *ModelCreateCommand) Handle(ctx console.Context) error {
	tableName := ctx.Option("table")
	if tableName == "" {
		ctx.Info("-t table name parameter cannot be empty")
		return nil
	}
	schema := facades.App().MakeSchema().Connection(ctx.Option("database"))
	info := databaseInfo{}
	var err error
	if info.Tables, err = schema.GetTables(); err != nil {
		ctx.Error(err.Error())
		return nil
	}
	if len(info.Tables) > 0 {
		isExist := false
		for i := range info.Tables {
			if info.Tables[i].Name == tableName {
				isExist = true
				columns, err := schema.GetColumns(tableName)
				if err != nil {
					ctx.Error(fmt.Sprintf("Failed to get columns: %s", err.Error()))
					return err
				}
				err = receiver.CreateModelStruct(ctx, columns, tableName)
				return err
			}
		}
		if !isExist {
			ctx.Info(fmt.Sprintf("%s table does not exist", tableName))
		}
	}
	return nil
}
func (receiver *ModelCreateCommand) CreateModelStruct(ctx console.Context, columns []schema.Column, tableName string) error {
	pwd, _ := os.Getwd()
	fileUrl := filepath.Join(pwd, "app", "models", tableName+".go")
	if file.Exists(fileUrl) {
		confirm, _ := ctx.Confirm(fmt.Sprintf("%s model already exists,Is it covered?", tableName))
		if !confirm {
			return nil
		}
	}
	modelStruct := &ModelStruct{}
	modelStruct.PackageName("models").ColumnGoField(columns).AllPkg()
	if err := file.Create(fileUrl, receiver.populateStub(receiver.getStub(), camelCase(tableName), tableName, modelStruct)); err != nil {
		return err
	}

	ctx.Success(fmt.Sprintf("Model %s created successfully", tableName))
	return nil
}
func (receiver *ModelCreateCommand) populateStub(stub, structName, tableName string, modelStruct *ModelStruct) string {
	stub = strings.ReplaceAll(stub, "DummyCommand", structName)
	stub = strings.ReplaceAll(stub, "columnStruct", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", modelStruct.PkgName)
	stub = strings.ReplaceAll(stub, "TableNameStr", tableName)
	if len(modelStruct.Pkg) > 0 {
		importPkg := "import (\n"
		for _, pkg := range modelStruct.Pkg {
			importPkg += fmt.Sprintf("\"%s\"\n", pkg)
		}
		importPkg += ")"
		stub = strings.ReplaceAll(stub, "ImportPkg", importPkg)
	} else {
		stub = strings.ReplaceAll(stub, "ImportPkg", "")
	}
	if len(modelStruct.Fields) > 0 {
		structContent := ""
		for _, field := range modelStruct.Fields {
			tag := ""
			if len(field.Tag) > 0 {
				tag = fmt.Sprintf("`%s`", strings.Join(field.Tag, " "))
			}
			structContent += fmt.Sprintf("\t%s %s %s\n", field.Field, field.FieldType, tag)
		}
		stub = strings.ReplaceAll(stub, "StructContent", strings.TrimRight(structContent, "\n"))
		structColumnsType := ""
		structColumns := ""
		for _, column := range modelStruct.Columns {
			camelCaseColumn := camelCase(column)
			if camelCaseColumn == "Id" {
				camelCaseColumn = strings.ToUpper(camelCaseColumn)
			}
			structColumnsType += fmt.Sprintf("\t%s string\n", camelCaseColumn)
			structColumns += fmt.Sprintf("\t\t%s: \"%s\",\n", camelCaseColumn, column)
		}
		stub = strings.ReplaceAll(stub, "StructColunmsType", strings.TrimRight(structColumnsType, "\n"))
		stub = strings.ReplaceAll(stub, "StructColunms", strings.TrimRight(structColumns, "\n"))
	}
	return stub
}

func (receiver *ModelCreateCommand) getStub() string {
	return `package DummyPackage

ImportPkg

type DummyCommand struct {
StructContent
}
func (m *Menu) TableName() string {
	return "TableNameStr"
}
type columnStructColumnStruct struct {
StructColunmsType
}
func DummyCommandColumns() columnStructColumnStruct{
	return columnStructColumnStruct{
StructColunms
  }
}
`
}

// 转大驼峰
func camelCase(str string) string {
	var text string
	for _, p := range strings.Split(str, "_") {
		// 字段首字母大写的同时, 是否要把其他字母转换为小写
		switch len(p) {
		case 0:
		case 1:
			text += strings.ToUpper(p[0:1])
		default:
			if strings.ToLower(p) == "id" {
				text += strings.ToUpper(p)
			} else {
				text += strings.ToUpper(p[0:1]) + p[1:]
			}
		}
	}
	return text
}

// 转小驼峰
func smallCamelCase(str string) string {
	var text string
	for i, p := range strings.Split(str, "_") {
		// 字段首字母大写的同时, 是否要把其他字母转换为小写
		switch i {
		case 0:
			text = p
		default:
			text += strings.ToUpper(p[0:1]) + p[1:]
		}
	}
	return text
}

// InSlice 判断字符串是否在 slice 中。
func InSlice(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

// varchar(32) 转成 varchar
func delTypeLen(str string) string {
	if strings.Contains(str, "(") {
		return str[0:strings.Index(str, "(")]
	}
	return str
}

// 数据表字段映射成go结构体字段
type GoField struct {
	Field     string
	FieldType string
	Tag       []string
}

type ModelStruct struct {
	PkgName string    //包名
	Pkg     []string  // 引入的包名
	Fields  []GoField // 字段
	Columns []string
}

// 设置包名
func (m *ModelStruct) PackageName(pkgName string) *ModelStruct {
	m.PkgName = pkgName
	return m
}

// 表字段转GoField
func (m *ModelStruct) ColumnGoField(columns []schema.Column) *ModelStruct {
	fields := make([]GoField, 0)
	for i := range columns {
		var (
			key   = columns[i].Name
			value = columns[i].Type
		)
		jsonTag := fmt.Sprintf("json:\"%s\"", smallCamelCase(key))
		formTag := fmt.Sprintf("form:\"%s\"", smallCamelCase(key))
		tags := []string{jsonTag, formTag}
		// id 列 默认主键tag
		if strings.ToLower(key) == "id" {
			tags = append(tags, "gorm:\"primaryKey\"")
		}
		m.Columns = append(m.Columns, key)
		field := GoField{
			Field:     camelCase(key),
			FieldType: MysqlToGotype[delTypeLen(value)],
			Tag:       tags,
		}
		fields = append(fields, field)
	}
	IdIndex := -1
	CreatedAtIndex := -1
	UpdatedAtIndex := -1
	DeletedAtIndex := -1
	for i, item := range fields {
		if item.Field == "ID" {
			IdIndex = i
		} else if item.Field == "CreatedAt" {
			CreatedAtIndex = i
		} else if item.Field == "UpdatedAt" {
			UpdatedAtIndex = i
		} else if item.Field == "DeletedAt" {
			DeletedAtIndex = i
		}
	}
	if DeletedAtIndex > -1 {
		fields[DeletedAtIndex] = GoField{Field: "orm.SoftDeletes"}
	}
	if CreatedAtIndex > -1 && UpdatedAtIndex > -1 && IdIndex > -1 {
		fields[IdIndex] = GoField{Field: "orm.Model"}
		fields = append(fields[:CreatedAtIndex], fields[CreatedAtIndex+1:]...)
		fields = append(fields[:UpdatedAtIndex-1], fields[UpdatedAtIndex:]...)
	} else if CreatedAtIndex > 0 && UpdatedAtIndex > 0 {
		fields = append(fields[:CreatedAtIndex], fields[CreatedAtIndex+1:]...)
		fields[UpdatedAtIndex-1] = GoField{Field: "orm.Timestamps"}
	}
	m.Fields = fields
	return m
}

// 添加所有字段依赖包
func (m *ModelStruct) AllPkg() *ModelStruct {
	for _, field := range m.Fields {
		if strings.Contains(field.FieldType, "carbon") {
			m.AddPkg("github.com/goravel/framework/support/carbon")
		}
		if strings.Contains(field.Field, "orm.") {
			m.AddPkg("github.com/goravel/framework/database/orm")
		}
	}
	return m
}
func (m *ModelStruct) AddPkg(pkg string) *ModelStruct {
	if !InSlice(m.Pkg, pkg) {
		m.Pkg = append(m.Pkg, pkg)
	}
	return m
}
