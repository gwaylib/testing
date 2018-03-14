package database

import (
	"database/sql"
	"strings"

	"github.com/gwaylib/errors"
)

// 添加一条数据，需要对象标注字段名
func AddObj(db Execer, tbName string, obj interface{}) (sql.Result, error) {
	return nil, nil
}

// 查询一个对像，可以是数组
func ScanObj(rows *sql.Rows, obj interface{}) error {
	return nil
}

// 执行一个通用的数字查询
func QueryInt(db Queryer, querySql string, args ...interface{}) (int64, error) {
	if len(querySql) == 0 {
		return 0, nil
	}
	num := sql.NullInt64{}
	if err := db.QueryRow(querySql, args...).Scan(&num); err != nil {
		if strings.Index(err.Error(), "Error 1146") != -1 {
			// no table for mysql
			return 0, nil
		}
		return 0, errors.As(err, querySql, args)
	}
	return num.Int64, nil
}

// 执行一个通用的字符查询
func QueryStr(db Queryer, querySql string, args ...interface{}) (string, error) {
	if len(querySql) == 0 {
		return "", nil
	}
	str := DBData("")
	if err := db.QueryRow(querySql, args...).Scan(&str); err != nil {
		if strings.Index(err.Error(), "Error 1146") != -1 {
			// no table for mysql
			return "", nil
		}
		return "", errors.As(err, querySql, args)
	}
	return str.String(), nil
}

// 执行一个通用的查询
// 因需要查标题，相对标准sql会慢一些，适用于偷懒查询的方式
// 即使发生错误返回至少是零长度的值
func QueryTable(db Queryer, querySql string, args ...interface{}) (titles []string, result [][]interface{}, err error) {
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
func QueryMap(db Queryer, querySql string, args ...interface{}) ([]map[string]interface{}, error) {
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
