/*
## 快速查询
### 单个查询
mdb := GetDB("./datastore.cfg", "master")
result, err := mdb.QueryOne("SELECT * FROM a WHERE id = ?", id)
// ...
// 或者
result, err := mdb.QueryData("SELECT * FROM a WHERE id = ?", id)
// ..

### 计数查询
mdb := GetDB("./datastore.cfg", "master")
total, err := mdb.QueryCount("SELECT count(*) FROM a WHERE id = ?", id)
// ...


### 批量查询
mdb := GetDB("./datastore.cfg", "master")

var (
	userInfoQsql = &qsql.Template{
		CountSql: `
SELECT
    count(1)
FROM
    %s
WHERE
    mobile = ?
`,
		DataSql: `
SELECT
    mobile "手机号"
FROM
    %s
WHERE
    mobile = ?
ORDER BY
    mobile
LIMIT ?, ? -- 须带有这两个参数
`,
	}
)

// 表格方式查询
total, titles, result, err := mdb.QueryList(
	userInfoQsql.FmtTempate("user_info_200601"),
    currPage*10, 10,
	"13800130000")
if err != nil {
	if !errors.ErrNoData.Equal(err) {
		log.Debug(errors.As(err, mobile, pid, uid))
		return c.String(500, "系统错误")
	}
	// 空数据
}


// 或者对像方式查询
total, titles, result, err := mdb.QueryMap(
	userInfoQsql.FmtTempate("user_info_200601"),
    currPage*10, 10,
	"13800130000")
if err != nil {
	if !errors.ErrNoData.Equal(err) {
		log.Debug(errors.As(err, mobile, pid, uid))
		return c.String(500, "系统错误")
	}
	// 无数据
}
*/
package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/gwaylib/errors"
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

// 执行一个通用的计数
func (db *DB) QueryCount(querySql string, args ...interface{}) (int, error) {
	if len(querySql) == 0 {
		return 0, nil
	}
	count := 0
	if err := db.QueryRow(querySql, args...).Scan(&count); err != nil {
		if strings.Index(err.Error(), "Error 1146") != -1 {
			// no table for mysql
			return count, nil
		}
		return 0, errors.As(err, querySql, args)
	}
	return count, nil
}

// 执行一个通用的查询
// 因需要查标题，相对标准sql会慢一些，适用于偷懒查询的方式
// 即使发生错误返回至少是零长度的值
func (db *DB) QueryData(querySql string, args ...interface{}) (titles []string, result [][]interface{}, err error) {
	titles = []string{}
	result = [][]interface{}{}
	rows, err := db.Query(querySql, args...)
	if err != nil {
		// mysql 1146: table no found
		if strings.Index(err.Error(), "1146") != -1 {
			// no table for mysql
			return titles, result, errors.ErrNoData.As(err)
		}
		return titles, result, errors.As(err, querySql, args)
	}
	defer rows.Close()

	titles, err = rows.Columns()
	if err != nil {
		return titles, result, errors.As(err, querySql, args)
	}

	for rows.Next() {
		r := makeDBData(len(titles))
		if err := rows.Scan(r...); err != nil {
			return titles, result, errors.As(err, querySql, args)
		}
		result = append(result, r)
	}
	if len(result) == 0 {
		return titles, result, errors.ErrNoData.As(args)
	}

	return titles, result, nil
}

// 查询一条数据，并发map结构返回，以便页面可以直接调用
// 因需要查标题，相对标准sql会慢一些，适用于偷懒查询的方式
// 即使发生错误返回至少是零长度的值
func (db *DB) QueryObject(querySql string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(querySql, args...)
	if err != nil {
		if strings.Index(err.Error(), "Error 1146") != -1 {
			// no table
			return []map[string]interface{}{}, errors.ErrNoData.As(err, args)
		}
		return []map[string]interface{}{}, errors.As(err, querySql, args)
	}
	defer rows.Close()

	// 列名
	names, err := rows.Columns()
	if err != nil {
		return []map[string]interface{}{}, errors.As(err, querySql, args)
	}

	// 取一条数据
	result := []map[string]interface{}{}
	for rows.Next() {
		r := makeDBData(len(names))
		if err := rows.Scan(r...); err != nil {
			return []map[string]interface{}{}, errors.As(err, querySql, args)
		}
		result := map[string]interface{}{}
		for i, name := range names {
			// 校验列名重复性
			_, ok := result[name]
			if ok {
				return []map[string]interface{}{}, errors.New("Already exist column name").As(querySql, name)
			}
			result[name] = r[i]
		}
	}
	if len(result) == 0 {
		return []map[string]interface{}{}, errors.ErrNoData.As(err, args)
	}
	return result, nil
}

// 读取列表表格数据
//
// 因需要查标题，相对标准sql会慢一些，适用于页面分页偷懒查询的方式
// 即使发生错误返回至少是零长度的值
//
// 返回
// int -- CountSql的数量
// []string -- DataSql的列名
// [][]interface{} -- DataSql的数据
func (db *DB) QueryList(tpl *Template, offset, limit int, args ...interface{}) (int, []string, [][]interface{}, error) {
	count, err := db.QueryCount(tpl.CountSql, args...)
	if err != nil {
		return 0, []string{}, [][]interface{}{}, errors.As(err, *tpl, offset, limit, args)
	}
	if count == 0 {
		return 0, []string{}, [][]interface{}{}, errors.ErrNoData.As(offset, limit, args)
	}
	args = append(args, offset)
	args = append(args, limit)
	titles, data, err := db.QueryData(tpl.DataSql, args...)
	if err != nil {
		return 0, []string{}, [][]interface{}{}, errors.As(err)
	}
	return count, titles, data, nil
}

// 读取列表对象数据
//
// 因需要查标题，相对标准sql会慢一些，适用于页面分页偷懒查询的方式
// 即使发生错误返回至少是零长度的值

// 返回
// int -- 不含limit的总数量
// []map[string]inteface{} -- 以title为key的数据
func (db *DB) QueryMap(tpl *Template, offset, limit int, args ...interface{}) (int, []map[string]interface{}, error) {
	count, err := db.QueryCount(tpl.CountSql, args)
	if err != nil {
		return 0, []map[string]interface{}{}, errors.As(err, *tpl, offset, limit, args)
	}
	if count == 0 {
		return 0, []map[string]interface{}{}, errors.ErrNoData.As(offset, limit, args)
	}
	args = append(args, offset)
	args = append(args, limit)
	result, err := db.QueryObject(tpl.DataSql, args...)
	if err != nil {
		return 0, []map[string]interface{}{}, errors.As(err)
	}
	return count, result, nil
}
