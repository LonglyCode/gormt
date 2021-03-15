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

type GormOptionFunc func(*gorm.DB) *gorm.DB

type Cache interface {
	Query(ctx context.Context, key string, value interface{}, query func(ctx context.Context) error) error
	Exec(ctx context.Context, key string, exec func(ctx context.Context) error) error
}

`

	genlogic = `

	{{$obj := .}}{{$list := $obj.Em}}
type {{$obj.StructName}}Mgr struct {
	DB *gorm.DB
	Cache cache.Cache
}

var {{$obj.StructName}}Set = wire.NewSet(wire.Struct(new({{$obj.StructName}}Mgr), "*"))

// GetTableName get sql table name.获取数据库名字
func (obj *{{$obj.StructName}}Mgr) TableName() string {
	return "{{$obj.TableName}}"
}

func (obj *{{$obj.StructName}}Mgr) PreTableName(s string) string {
	b := strings.Builder{}
	b.WriteString(obj.TableName())
	b.WriteString(".")
	b.WriteString(s)
	return b.String()
}

 {{range $ofm := $obj.Primay}}
	// Updates 更新
	func (obj *{{$obj.StructName}}Mgr) Updates(ctx context.Context, input *{{$obj.StructName}}, q *{{CapLowercase $obj.StructName}}Q) error {
		Q := q.Query(ctx)
		key := ""
		if k, _ := Q.InstanceGet("cache_key"); k != nil {
			key = utils.AsString(k)
		}
		return obj.Cache.Exec(ctx, key, func(ctx context.Context) error {
			return Q.Updates(*input).Error
		})
	}

	// Delete By ID
	func (obj *{{$obj.StructName}}Mgr) Delete(ctx context.Context, {{GenFListIndex $ofm 2}}) error {
		err := obj.DB.WithContext(ctx).Delete(&{{$obj.StructName}}{}, {{GenFListIndex $ofm 4}}).Error
		return err
	}
	// CacheKey{{GenFListIndex $ofm 5}} CacheKey generate by ids	
	func (obj *{{$obj.StructName}}Mgr) CacheKey{{GenFListIndex $ofm 5}}({{GenFListIndex $ofm 2}}) string {
		return strings.Join([]string{obj.TableName(), "{{GenFListIndex $ofm 4}}", utils.AsString({{GenFListIndex $ofm 4}})}, "_")
	}
{{end}}

// create 创建
func (obj *{{$obj.StructName}}Mgr) Create(ctx context.Context, input *{{$obj.StructName}}) (*{{$obj.StructName}}, error) {
	if err := obj.DB.WithContext(ctx).Create(input).Error; err != nil {
		return nil, err
	}
	return input, nil
}


type {{CapLowercase $obj.StructName}}Q struct {
	Mgr  *{{$obj.StructName}}Mgr
	opts []GormOptionFunc
}

func (obj *{{$obj.StructName}}Mgr) Q() *{{CapLowercase $obj.StructName}}Q {
	return &{{CapLowercase $obj.StructName}}Q{
		Mgr: obj,
	}
}

// QueryDefault 查询列表 
func (obj *{{CapLowercase $obj.StructName}}Q) List(ctx context.Context, value interface{}) (int64, error) {
	var cnt int64
	obj.Query(ctx).Offset(-1).Find(value).Count(&cnt)
	err := obj.Query(ctx).Order("update_time desc").Find(value).Error
	return cnt, err
}

//QueryDefault 查询单个
func (obj *{{CapLowercase $obj.StructName}}Q) One(ctx context.Context, value interface{}) error {
	Q := obj.Query(ctx)
	cache_key := ""
	// 缓存只适用于单行全列的情况
	if k, _ := Q.Get("cache_key"); len(obj.opts) == 1 && k != nil {
		cache_key = utils.AsString(k)
		Q = Q.Select("*")
	}
	return obj.Mgr.Cache.Query(ctx, cache_key, value, func(ctx context.Context) error {
		err := Q.First(value).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrIdCanNotFound
		}
		return err
	})
}

func (obj *{{CapLowercase $obj.StructName}}Q) Query(ctx context.Context) *gorm.DB {
	db := obj.Mgr.DB.WithContext(ctx).Model(&{{$obj.StructName}}{})
	for _, f := range obj.opts {
		db = f(db)
	}
	return db
}

func (obj *{{CapLowercase $obj.StructName}}Q) Select(strings ...string) *{{CapLowercase $obj.StructName}}Q {
	fn := func(db *gorm.DB) *gorm.DB {
		if len(strings) > 0 {
			var ss []string
			for _, s := range strings {
				ss = append(ss, obj.Mgr.PreTableName(s))
			}
			db = db.Select(ss)
		}
		return db
	}
	obj.opts = append(obj.opts, fn)
	return obj
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
// {{$oem.ColStructName}} {{$oem.ColName}}获取 {{$oem.Notes}}
func (obj *{{CapLowercase $obj.StructName}}Q) {{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) *{{CapLowercase $obj.StructName}}Q {
	fn := func(db *gorm.DB) *gorm.DB {
		db = db.Where(obj.Mgr.PreTableName("{{$oem.ColName}} = ?"), {{CapLowercase $oem.ColStructName}})
		return db
	} 
	obj.opts = append(obj.opts, fn)
	return obj
}
{{$t := HasSuffix $oem.ColStructName "Time"}}
{{$id := HasSuffix $oem.ColStructName "ID"}}
{{$str := IsType $oem.Type "string"}}
{{if $t}}
// {{$oem.ColStructName}}Interval {{$oem.ColName}}获取时间区间 {{$oem.Notes}}
func (obj *{{CapLowercase $obj.StructName}}Q) {{$oem.ColStructName}}Interval(interval []interface{}) *{{CapLowercase $obj.StructName}}Q {
	fn := func(db *gorm.DB) *gorm.DB {
		db = db.Where(obj.Mgr.PreTableName("{{$oem.ColName}} between ? and ?"), interval[0], interval[1])
		return db
	}
	obj.opts = append(obj.opts, fn)
	return obj
}
{{else if $id}}
// {{$oem.ColStructName}}In {{$oem.ColName}}获取in {{$oem.Notes}}
func (obj *{{CapLowercase $obj.StructName}}Q) {{$oem.ColStructName}}In({{CapLowercase $oem.ColStructName}}s ...{{$oem.Type}}) *{{CapLowercase $obj.StructName}}Q {
	fn := func(db *gorm.DB) *gorm.DB {
		if len({{CapLowercase $oem.ColStructName}}s) > 0 {
			db = db.Where(obj.Mgr.PreTableName("{{$oem.ColName}} in (?)"), {{CapLowercase $oem.ColStructName}}s)
		}
		return db
	} 
	obj.opts = append(obj.opts, fn)
	return obj
}
{{else if $str}}
// {{$oem.ColStructName}}Like {{$oem.ColName}}获取like {{$oem.Notes}}
func (obj *{{CapLowercase $obj.StructName}}Q) {{$oem.ColStructName}}Like({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) *{{CapLowercase $obj.StructName}}Q {
	fn := func(db *gorm.DB) *gorm.DB {
		db = db.Where(obj.Mgr.PreTableName("{{$oem.ColName}} like ?"), "%"+{{CapLowercase $oem.ColStructName}}+"%")
		return db
	}
	obj.opts = append(obj.opts, fn)
	return obj
}
{{end}}
{{end}}

func (obj *{{CapLowercase $obj.StructName}}Q) Pagination(para *{{$obj.StructName}}ReqParams) *{{CapLowercase $obj.StructName}}Q {
	fn := func(db *gorm.DB) *gorm.DB {
		if para.PageNum > 0 && para.PageSize > 0 {
			if para.PageSize > 1000 {
				// 不允许大于1000量
				db = db.Limit(1000)
			} else {
				db = db.Limit(para.PageSize).Offset((para.PageNum - 1) * para.PageSize)
			}
		} else if !para.Export {
			// 非导出，也是限制 1000
			db = db.Limit(1000)
		}
		return db
	}
	obj.opts = append(obj.opts, fn)
	return obj
}

func (opt *{{CapLowercase $obj.StructName}}Q) Filter(para *{{$obj.StructName}}ReqParams) *{{CapLowercase $obj.StructName}}Q {
	if para != nil {
		opt.Select(para.Fields...)
		opt.Pagination(para)	
		if para.Query != nil {
		{{range $oem := $obj.Em}}
			{{$t := HasSuffix $oem.ColStructName "Time"}}
			{{$id := HasSuffix $oem.ColStructName "ID"}}
			{{$str := IsType $oem.Type "string"}}
			{{if $str}}
			if para.Query.{{$oem.ColStructName}} != "" {
				opt.{{$oem.ColStructName}}(para.Query.{{$oem.ColStructName}})
			} 
			if para.Query.{{$oem.ColStructName}}Like != "" {
				opt.{{$oem.ColStructName}}Like(para.Query.{{$oem.ColStructName}}Like)
			} 
			{{else}}
			if para.Query.{{$oem.ColStructName}} != 0 {
				opt.{{$oem.ColStructName}}(para.Query.{{$oem.ColStructName}})
			} 
			{{end}}
			{{if $t}} 
			if len(para.Query.{{$oem.ColStructName}}Interval) > 0 {
				opt.{{$oem.ColStructName}}Interval(para.Query.{{$oem.ColStructName}}Interval)
			}
			{{end}}
			{{if $id}} 
			if len(para.Query.{{$oem.ColStructName}}In) > 0 {
				opt.{{$oem.ColStructName}}In(para.Query.{{$oem.ColStructName}}In...)
			}
			{{end}}

		{{end}}
		}
	}
	return opt
}
 //////////////////////////primary index case ////////////////////////////////////////////
`

	genPreload      = ``
	genPreloadMulti = ``
)
