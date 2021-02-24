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
	glog "gorm.io/gorm/logger"
)

var globalIsRelated bool = true  // 全局预加载

type GormOptionFunc func(*gorm.DB) *gorm.DB

// prepare for other
type _BaseMgr struct {
	*gorm.DB
}

// GetDB get gorm.DB info
func (obj *_BaseMgr) GetDB() *gorm.DB {
	return Session(obj.DB, ctx)
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

type scanObj struct{}

// NewScanObj 创建返回
func NewScanObj(ctx context.Context, o interface{}) context.Context {
	return context.WithValue(ctx, scanObj{}, o)
}

// GetScanObjCtx 从上下文中获取跟踪ID
func GetScanObjCtx(ctx context.Context) interface{} {
	v := ctx.Value(scanObj{})
	if v != nil {
		return v
	}
	return nil
}
`

	genlogic = `

	{{$obj := .}}{{$list := $obj.Em}}
type {{$obj.StructName}}Mgr struct {
	DB *gorm.DB
	Cache *cache.Cache
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
	func (obj *{{$obj.StructName}}Mgr) Updates(ctx context.Context, input *{{$obj.StructName}}, opt GormOptionFunc) error {
		Q := obj.query(ctx, opt)
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

// QueryDefault 查询列表 
func (obj *{{$obj.StructName}}Mgr) QueryDefault(ctx context.Context, opts ...GormOptionFunc) ([]*{{$obj.StructName}}, int64, error) {
	var (
		list []*{{$obj.StructName}}
		cnt  int64
	)
	// for count
	Q := obj.query(ctx, opts...)
	Q.Offset(-1).Find(&list).Count(&cnt)
	// fore list
	Q = obj.query(ctx, opts...)
	err := Q.Order("update_time desc").Find(&list).Error
	return list, cnt, err
}

//QueryDefault 查询单个
func (obj *{{$obj.StructName}}Mgr) QueryOne(ctx context.Context, value interface{}, opts ...GormOptionFunc) error {
	Q := obj.query(ctx, opts...)
	cache_key := ""
	// 缓存只适用于单行全列的情况
	if k, _ := Q.Get("cache_key"); len(opts) == 1 && k != nil {
		cache_key = utils.AsString(k)
		Q = Q.Select("*")
	}
	return obj.Cache.Query(ctx, cache_key, value, func(ctx context.Context) error {
		err := Q.First(value).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrIdCanNotFound
		}
		return err
	})
}

func (obj *{{$obj.StructName}}Mgr) query(ctx context.Context, opts ...GormOptionFunc) *gorm.DB {
	db := obj.DB.WithContext(ctx).Model(&{{$obj.StructName}}{})
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
				if para.PageNum*para.PageSize > 1000 {
					// 不允许大于1000量
					db = db.Limit(1000)
				} else {
					db = db.Limit(para.PageSize).Offset((para.PageNum - 1) * para.PageSize)
				}
			} else if !para.Export {
				// 非导出，也是限制 1000
				db = db.Limit(1000)
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
