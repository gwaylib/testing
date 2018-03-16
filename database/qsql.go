package database

import (
	"database/sql"
	"fmt"

	"github.com/gwaylib/errors"
	"github.com/jmoiron/sqlx"
)

// 自增回调接口
type AutoIncrAble interface {
	// notify for last id
	SetLastInsertId(id int64, err error)
}

// 执行器
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// 查询器
type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// 扫描器
type Scaner interface {
	Close() error
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(...interface{}) error
}

const (
	addObjSql = `
INSERT INTO %s
	(%s)
VALUES
	(%s)
	`
)

// 添加一条数据，需要结构体至少标注字段名 `db:"name"`, 标签详情请参考github.com/jmoiron/sqlx
// 关于drvNames的设计说明
// 因支持一个可变参数, 或未填，将使用默认值:DEFAULT_DRV_NAME
func insertStruct(exec Execer, obj interface{}, tbName string, drvNames ...string) (sql.Result, error) {
	drvName := DEFAULT_DRV_NAME
	drvNamesLen := len(drvNames)
	if drvNamesLen > 0 {
		if drvNamesLen != 0 {
			panic(errors.New("'drvNames' expect only one argument").As(drvNames))
		}
		drvName = drvNames[0]
	}
	names, inputs, vals, err := reflectInsertStruct(obj, drvName)
	if err != nil {
		return nil, errors.As(err)
	}
	execSql := fmt.Sprintf(tbName, names, inputs)
	result, err := exec.Exec(execSql, vals...)
	if err != nil {
		return nil, errors.As(err)
	}
	incr, ok := obj.(AutoIncrAble) // need obj is ptr kind.
	if ok {
		incr.SetLastInsertId(result.LastInsertId())
	}
	return result, nil
}

// 扫描结果到一个结构体，该结构体可以是数组
// 代码设计请参阅github.com/jmoiron/sqlx
func scanStruct(rows Scaner, obj interface{}) error {
	if err := sqlx.StructScan(rows, obj); err != nil {
		return errors.As(err)
	}
	return nil
}

// 查询一个对象
func queryStruct(db Queryer, obj interface{}, querySql string, args ...interface{}) error {
	rows, err := db.Query(querySql, args...)
	if err != nil {
		return errors.As(err, querySql, args)
	}
	defer Close(rows)

	if err := scanStruct(rows, obj); err != nil {
		return errors.As(err, querySql, args)
	}

	return nil
}

// 执行一个通用的数字查询
func queryInt(db Queryer, querySql string, args ...interface{}) (int64, error) {
	if len(querySql) == 0 {
		return 0, nil
	}
	num := sql.NullInt64{}
	if err := db.QueryRow(querySql, args...).Scan(&num); err != nil {
		return 0, errors.As(err, querySql, args)
	}
	return num.Int64, nil
}

// 执行一个通用的字符查询
func queryStr(db Queryer, querySql string, args ...interface{}) (string, error) {
	if len(querySql) == 0 {
		return "", nil
	}
	str := DBData("")
	if err := db.QueryRow(querySql, args...).Scan(&str); err != nil {
		return "", errors.As(err, querySql, args)
	}
	return str.String(), nil
}

// 执行一个通用的查询
// 因需要查标题，相对标准sql会慢一些，适用于偷懒查询的方式
// 即使发生错误返回至少是零长度的值
func queryTable(db Queryer, querySql string, args ...interface{}) (titles []string, result [][]interface{}, err error) {
	titles = []string{}
	result = [][]interface{}{}
	rows, err := db.Query(querySql, args...)
	if err != nil {
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
func queryMap(db Queryer, querySql string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(querySql, args...)
	if err != nil {
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
