package genfunc

const (
	genTnf = `
// TableName get sql table name.获取数据库表名
func (m *{{.StructName}}) TableName() string {
	return "{{.TableName}}"
}
`
	genColumn = `
// {{.StructName}}Columns get sql column name.获取数据库列名
var {{.StructName}}Columns = struct { {{range $em := .Em}}
	{{$em.StructName}} string{{end}}    
	}{ {{range $em := .Em}}
		{{$em.StructName}}:"{{$em.ColumnName}}",  {{end}}           
	}
`
	genBase = `
package model
import (
	"gorm.io/gorm"
)

var globalIsRelated bool = true  // 全局预加载

type GormOptionFunc func(*gorm.DB) *gorm.DB

// prepare for other
type _BaseMgr struct {
	*gorm.DB
}

// GetDB get gorm.DB info
func (obj *_BaseMgr) GetDB() *gorm.DB {
	return obj.DB
}

type options struct {
	query map[string]interface{}
}

// Option overrides behavior of Connect.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}
	`

	genlogic = `

	{{$obj := .}}{{$list := $obj.Em}}
type {{$obj.StructName}}Mgr struct {
	DB *gorm.DB
}

var {{$obj.StructName}}Set = wire.NewSet(wire.Struct(new({{$obj.StructName}}Mgr), "*"))

// GetTableName get sql table name.获取数据库名字
func (obj *{{$obj.StructName}}Mgr) TableName() string {
	return "{{$obj.TableName}}"
}

func (obj *{{$obj.StructName}}Mgr) PreTableName(s string) string {
	return fmt.Sprintf("%s.%s", obj.TableName(), s)
}

 {{range $ofm := $obj.Index}}
{{$primary := IsPrimary $ofm}}
{{if $primary}}
	// Get 获取
	func (obj *{{$obj.StructName}}Mgr) Get({{GenFListIndex $ofm 2}}) (*{{$obj.StructName}}, error) {
		result := &{{$obj.StructName}}{}	
		err := obj.DB.Where("{{GenFListIndex $ofm 3}}", {{GenFListIndex $ofm 4}}).First(&result).Error
		return result, err
	}
	// Updates 更新
	func (obj *{{$obj.StructName}}Mgr) Updates({{GenFListIndex $ofm 2}}, column string, value interface{}) error {
	if {{GenFListIndex $ofm 3}} == 0 {
		return errors.New("id不能为空")
	}
	m := &{{$obj.StructName}}{
		{{GenFListIndex $ofm 4}}: {{GenFListIndex $ofm 4}},
	}
	return obj.DB.Model(m).Update(column, value).Error
}
{{end}}
{{end}}

// create 创建
func (obj *{{$obj.StructName}}Mgr) Create(input *{{$obj.StructName}}) (*{{$obj.StructName}}, error) {
	if err := obj.DB.Create(input).Error; err != nil {
		return nil, err
	}
	return input, nil
}


func (obj *{{$obj.StructName}}Mgr) Save(input *{{$obj.StructName}}) error {
	return obj.DB.Model(input).Updates(*input).Error
}

// QueryDefault 查询列表 
func (obj *{{$obj.StructName}}Mgr) QueryDefault(ctx context.Context, opts ...GormOptionFunc) ([]*{{$obj.StructName}}, int64, error) {
	var (
		list []*{{$obj.StructName}}
		cnt  int64
	)
	// for count
	Q := obj.query(obj.DB.WithContext(ctx), opts...)
	Q.Offset(-1).Find(&list).Count(&cnt)
	// fore list
	Q = obj.query(obj.DB.WithContext(ctx), opts...)
	err := Q.Order("update_time desc").Find(&list).Error
	return list, cnt, err
}

//QueryDefault 查询单个
func (obj *{{$obj.StructName}}Mgr) QueryOne(ctx context.Context, opts ...GormOptionFunc) (*{{$obj.StructName}}, bool, error) {
	one := &{{$obj.StructName}}{}
	err := obj.query(obj.DB.WithContext(ctx), opts...).First(one).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}
	return one, true, nil
}

func (obj *{{$obj.StructName}}Mgr) query(db *gorm.DB, opts ...GormOptionFunc) *gorm.DB {
	for _, f := range opts {
		db = f(db)
	}
	return db
}

func (obj *{{$obj.StructName}}Mgr) WithSelect(strings ...string) GormOptionFunc {
	return func(db *gorm.DB) *gorm.DB {
		if len(strings) > 0 {
			var ss []string
			for _, s := range strings {
				ss = append(ss, obj.PreTableName(s))
			}
			db = db.Select(ss)
		}
		return db
	}
}

type {{$obj.StructName}}ReqParams struct {
	Query    *{{$obj.StructName}}Params {{JsonStr "query"}}
	Export   bool        {{JsonStr "export"}}
	Fields   []string    {{JsonStr "fields"}}
	PageNum  int         {{JsonStr "page_num"}} 
	PageSize int         {{JsonStr "page_size"}}           
}

type {{$obj.StructName}}Params struct {
	{{range $oem := $obj.Em}}
		{{$t := HasSuffix $oem.ColStructName "Time"}}
		{{$id := HasSuffix $oem.ColStructName "ID"}}
		{{$str := IsType $oem.Type "string"}}

		{{$oem.ColStructName}} {{$oem.Type}} {{JsonStr $oem.ColName}} 	

		{{if $str}}
		{{$oem.ColStructName}}Like {{$oem.Type}} {{JsonStr (print $oem.ColName "|like")}} 	
		{{end}}

		{{if $t}} 
		{{$oem.ColStructName}}Interval []interface{} {{JsonStr (print $oem.ColName "|interval")}} 	
		{{end}}

		{{if $id}} 
		{{$oem.ColStructName}}In []{{$oem.Type}} {{JsonStr (print $oem.ColName "|in")}} 	
		{{end}}
	{{end}}
}


//////////////////////////option case ////////////////////////////////////////////
{{range $oem := $obj.Em}}
// With{{$oem.ColStructName}} {{$oem.ColName}}获取 {{$oem.Notes}}
func (obj *{{$obj.StructName}}Mgr) With{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) GormOptionFunc {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Where(obj.PreTableName("{{$oem.ColName}} = ?"), {{CapLowercase $oem.ColStructName}})
		return db
	} 
}
{{$t := HasSuffix $oem.ColStructName "Time"}}
{{$id := HasSuffix $oem.ColStructName "ID"}}
{{$str := IsType $oem.Type "string"}}
{{if $t}}
// With{{$oem.ColStructName}}Interval {{$oem.ColName}}获取时间区间 {{$oem.Notes}}
func (obj *{{$obj.StructName}}Mgr) With{{$oem.ColStructName}}Interval(interval []interface{}) GormOptionFunc {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Where(obj.PreTableName("{{$oem.ColName}} between ? and ?"), interval[0], interval[1])
		return db
	}
}
{{else if $id}}
// With{{$oem.ColStructName}}In {{$oem.ColName}}获取in {{$oem.Notes}}
func (obj *{{$obj.StructName}}Mgr) With{{$oem.ColStructName}}In({{CapLowercase $oem.ColStructName}}s ...{{$oem.Type}}) GormOptionFunc {
	return func(db *gorm.DB) *gorm.DB {
		if len({{CapLowercase $oem.ColStructName}}s) > 0 {
			db = db.Where(obj.PreTableName("{{$oem.ColName}} in (?)"), {{CapLowercase $oem.ColStructName}}s)
		}
		return db
	} 
}
{{else if $str}}
// With{{$oem.ColStructName}}Like {{$oem.ColName}}获取like {{$oem.Notes}}
func (obj *{{$obj.StructName}}Mgr) With{{$oem.ColStructName}}Like({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) GormOptionFunc {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Where(obj.PreTableName("{{$oem.ColName}} like ?"), "%"+{{CapLowercase $oem.ColStructName}}+"%")
		return db
	}
}
{{end}}
{{end}}

func (opt *{{$obj.StructName}}Mgr) Filter(para *{{$obj.StructName}}ReqParams) GormOptionFunc {
	return func(db *gorm.DB) *gorm.DB {
		if para != nil {
			db = db.Scopes(opt.WithSelect(para.Fields...))
			if para.PageNum > 0 && para.PageSize > 0 {
				db = db.Limit(para.PageSize).Offset((para.PageNum - 1) * para.PageSize)
			}
			if para.Query != nil {
			{{range $oem := $obj.Em}}
				{{$t := HasSuffix $oem.ColStructName "Time"}}
				{{$id := HasSuffix $oem.ColStructName "ID"}}
				{{$str := IsType $oem.Type "string"}}
				{{if $str}}
				if para.Query.{{$oem.ColStructName}} != "" {
					db = db.Scopes(opt.With{{$oem.ColStructName}}(para.Query.{{$oem.ColStructName}}))
				} 
				if para.Query.{{$oem.ColStructName}}Like != "" {
					db = db.Scopes(opt.With{{$oem.ColStructName}}Like(para.Query.{{$oem.ColStructName}}Like))
				} 
				{{else}}
				if para.Query.{{$oem.ColStructName}} != 0 {
					db = db.Scopes(opt.With{{$oem.ColStructName}}(para.Query.{{$oem.ColStructName}}))
				} 
				{{end}}
				{{if $t}} 
				if len(para.Query.{{$oem.ColStructName}}Interval) > 0 {
					db = db.Scopes(opt.With{{$oem.ColStructName}}Interval(para.Query.{{$oem.ColStructName}}Interval))
				}
				{{end}}
				{{if $id}} 
				if len(para.Query.{{$oem.ColStructName}}In) > 0 {
					db = db.Scopes(opt.With{{$oem.ColStructName}}In(para.Query.{{$oem.ColStructName}}In...))
				}
				{{end}}

			{{end}}
			}
		}
		return db
	}
}
 //////////////////////////primary index case ////////////////////////////////////////////
`

	genPreload      = ``
	genPreloadMulti = ``
)
