package qsql

import (
	"fmt"
	"time"
)

type DBData string

func (d *DBData) Scan(i interface{}) error {
	if i == nil {
		return nil
	}
	switch i.(type) {
	case int64:
		*d = DBData(fmt.Sprintf("%d", i))
	case float64:
		*d = DBData(fmt.Sprint(i))
	case []byte:
		*d = DBData(string(i.([]byte)))
	case string:
		*d = DBData(i.(string))
	case bool:
		*d = DBData(fmt.Sprintf("%t", i))
	case time.Time:
		*d = DBData(i.(time.Time).Format("2006-01-02 15:04:05"))
	default:
		*d = DBData(fmt.Sprint(i))
	}
	return nil
}
func (d *DBData) String() string {
	return string(*d)
}

func makeDBData(l int) []interface{} {
	r := make([]interface{}, l)
	for i := 0; i < l; i++ {
		d := DBData("")
		r[i] = &d
	}
	return r
}

type Template struct {
	CountSql string // 读取数据总行数
	DataSql  string // 读取数据细节
}

// 返回一个fmt.Sprintf()格式化Sql后的Template，
// 主要用于分表的读取
func (t Template) FmtTemplate(args ...interface{}) *Template {
	countSql := t.CountSql
	if len(countSql) > 0 {
		countSql = fmt.Sprintf(t.CountSql, args...)
	}
	dataSql := t.DataSql
	if len(dataSql) > 0 {
		dataSql = fmt.Sprintf(t.DataSql, args...)
	}

	return &Template{
		CountSql: countSql,
		DataSql:  dataSql,
	}
}
