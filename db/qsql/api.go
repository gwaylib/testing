package qsql

import (
	"strings"

	"github.com/gwaylib/datastore"
	"github.com/gwaylib/errors"
)

// 读取数据
// 返回
// int -- CountSql的数量
// []string -- DataSql的列名
// [][]inteface{} -- DataSql的数据
func Query(db *datastore.DB, tpl *Template, args ...interface{}) (int, []string, [][]interface{}, error) {
	count, err := QueryCount(db, tpl.CountSql, args...)
	if err != nil {
		return 0, nil, nil, errors.As(err)
	}
	if count == 0 {
		return 0, nil, nil, errors.ErrNoData.As(args)
	}
	args = append(args, tpl.Offset)
	args = append(args, tpl.Limit)
	titles, data, err := QueryData(db, tpl.DataSql, args...)
	if err != nil {
		return 0, nil, nil, errors.As(err)
	}
	return count, titles, data, nil
}

// 读取数据
// 返回
// int -- 不含limit的总数量
// []map[string]inteface{} -- 数据
func QueryMap(db *datastore.DB, tpl *Template, args ...interface{}) (int, []map[string]interface{}, error) {
	count, err := QueryCount(db, tpl.CountSql, args)
	if err != nil {
		return 0, nil, errors.As(err)
	}
	if count == 0 {
		return 0, nil, errors.ErrNoData.As(args)
	}
	args = append(args, tpl.Offset)
	args = append(args, tpl.Limit)
	querySql := tpl.DataSql

	rows, err := db.Query(querySql, args...)
	if err != nil {
		// mysql 1146: table no found
		if strings.Index(err.Error(), "1146") != -1 {
			// no table
			return 0, nil, errors.ErrNoData.As(err)
		}
		return 0, nil, errors.As(err, querySql, args)
	}
	defer rows.Close()

	titles, err := rows.Columns()
	if err != nil {
		return 0, nil, errors.As(err, querySql, args)
	}

	result := []map[string]interface{}{}
	for rows.Next() {
		r := newDBData(len(titles))
		if err := rows.Scan(r...); err != nil {
			return 0, nil, errors.As(err, querySql, args)
		}
		value := map[string]interface{}{}
		for i, title := range titles {
			// 校验列名重复性
			_, ok := value[title]
			if ok {
				panic("Already exist column name:" + title)
			}
			value[title] = r[i]
		}
		result = append(result, value)
	}
	if len(result) == 0 {
		return 0, nil, errors.ErrNoData.As(querySql, args)
	}
	return count, result, nil
}

// 查询一条数据，并发map结构返回，以便页面可以直接调用
func QueryOne(db *datastore.DB, querySql string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := db.Query(querySql, args...)
	if err != nil {
		if strings.Index(err.Error(), "Error 1146") != -1 {
			// no table
			return nil, errors.ErrNoData.As(err)
		}
		return nil, errors.As(err, querySql, args)
	}
	defer rows.Close()

	// 列名
	names, err := rows.Columns()
	if err != nil {
		return nil, errors.As(err, querySql, args)
	}

	// 取一条数据
	for rows.Next() {
		r := newDBData(len(names))
		if err := rows.Scan(r...); err != nil {
			return nil, errors.As(err, querySql, args)
		}
		result := map[string]interface{}{}
		for i, name := range names {
			// 校验列名重复性
			_, ok := result[name]
			if ok {
				panic("Already exist column name:" + name)
			}
			result[name] = r[i]
		}
		return result, nil
	}
	return nil, errors.ErrNoData.As(args)
}

// 执行一个通用的计数
func QueryCount(db *datastore.DB, querySql string, args ...interface{}) (int, error) {
	if len(querySql) == 0 {
		return 1, nil
	}
	count := 0
	if err := db.QueryRow(querySql, args...).Scan(&count); err != nil {
		if strings.Index(err.Error(), "Error 1146") != -1 {
			// no table
			return count, nil
		}
		return 0, errors.As(err, querySql, args)
	}
	return count, nil
}

// 执行一个通用的查询
func QueryData(db *datastore.DB, querySql string, args ...interface{}) (titles []string, result [][]interface{}, err error) {
	titles = []string{}
	result = [][]interface{}{}
	rows, err := db.Query(querySql, args...)
	if err != nil {
		// mysql 1146: table no found
		if strings.Index(err.Error(), "1146") != -1 {
			// no table
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
		r := newDBData(len(titles))
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
