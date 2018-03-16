# 说明

因项目需要集成了些库，以方便使用

参考资料

标准库

    database/sql

sqlx框架

    https://github.com/jmoiron/sqlx

本项目仅实现新增与查询功能

未实现删与改的原因是个人认为标准库很适合使用了，且更新时需要注意数据库性能问题, 不需要另行封装;

未实现其他功能比如创建表等是个人认为专业的数据库工具更适用于这方面的操作与实现。

# 使用例子：

## 配置文件

配置文件(假定为:"./datastore.cfg")中配置如下格式:

``` text
# 主库
[master]
driver: mysql
dsn: username:passwd@tcp(127.0.0.1:3306)/center?timeout=30s&strict=true&loc=Local&parseTime=true&allowOldPasswords=1

# 日志库
[log]
driver: mysql
dsn: username:passwd@tcp(127.0.0.1:3306)/log?timeout=30s&strict=true&loc=Local&parseTime=true&allowOldPasswords=1
```

## 性能级别建议使用标准库以便可灵活运用
``` text
// 导入驱动库
import <database driver package>
import "github.com/gwaylib/datastore/database"
```

### 使用标准查询
``` text
mdb := database.CacheDB("./datastore.cfg", "master")
// or mdb = <sql.Tx>
// or mdb = <sql.Stmt>
row := database.QueryRow(mdb, "SELECT * ...")
// ...

rows, err := database.Query(mdb, "SELECT * ...")
// ...

result, err := database.Exec(mdb, "UPDATE ...")
// ...
```

### 快速新增数据
``` text
// 定义表结构体
type User struct{
    // autoincrement 标签在插入时将被自动忽略插入
    Id   int64 `db:"id,autoincrement"`
    Name string `db:"name"`
}

// 实现自增回调接口
// AutoIncrAble接口应配合autoincrement标签使用
func (u *User)SetLastInsertId(id int64, err error){
    if err != nil{
        panic(err)
    }
    u.Id = id
}

var u = &User{
    Name:"testing",
}

// 新增例子一：
// 在需要时设置默认驱动名
// database.DEFAULT_DRV_NAME = database.DRV_NAME_MYSQL
if _, err := database.InsertStruct(mdb, u, "testing"); err != nil{
    // ... 
}
// ...

// 新增例子二：
if _, err := database.InsertStruct(mdb, u, "testing", "mysql"); err != nil{
    // ... 
}
// ...
```

## 快速查询, 用于通用性的查询，例如js页面返回
### 查询结果到结构体
``` text
import <database driver package>
import "github.com/gwaylib/datastore/database"

// 定义表结构体
type User struct{
    Id   int64 `db:"id"`
    Name string `db:"name"`
}

// 方法一

mdb := database.CacheDB("./datastore.cfg", "master")
// or mdb = <sql.Tx>
// or mdb = <sql.Stmt>
var u = []*User{}
result, err := database.QueryStructs(mdb, &u, "SELECT id, name FROM a WHERE id = ?", id)
if err != nil{
    // ...
}
if len(u) == 0{
    // data not found
    // ...
}
// .. 

// 或者
// 方法二

mdb := database.CacheDB("./datastore.cfg", "master")
// or mdb = <sql.Tx>
// or mdb = <sql.Stmt>
rows, err := database.Query(mdb, "SELECT id, name FROM a WHERE id = ?", id)
if err != nil {
    // ...
}
defer database.Close(rows)
var u = []*User{}
if err := database.ScanStructs(rows, &u); err != nil{
    // ...
}
if len(u) == 0{
    // data not found
    // ...
}

```

### 查询结果到整型
```text
import <database driver package>
import "github.com/gwaylib/datastore/database"

mdb := database.CacheDB("./datastore.cfg", "master")
// or mdb = <sql.Tx>
// or mdb = <sql.Stmt>
count, err := database.QueryInt(mdb, "SELECT count(*) FROM a WHERE id = ?", id)
// ...
```

### 批量查询
```text
import <database driver package>
import "github.com/gwaylib/datastore/database"

mdb := database.CacheDB("./datastore.cfg", "master")

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
LIMIT ?, ?
`,
	}
)

// 表格方式查询
result, err := database.QueryTable(
    mdb,
    userInfoQsql.FmtTempate("user_info_200601").DataSql,
    "13800130000", currPage*10, 10)
if err != nil {
    if !errors.ErrNoData.Equal(err) {
        log.Debug(errors.As(err, mobile, pid, uid))
        return c.String(500, "系统错误")
    }
    // 空数据
}

// 或者对象方式查询
result, err := database.QueryMap(
    mdb,
    userInfoQsql.FmtTempate("user_info_200601").DataSql,
    "13800130000",
    currPage*10, 10) 
if err != nil {
    if !errors.ErrNoData.Equal(err) {
        log.Debug(errors.As(err, mobile, pid, uid))
        return c.String(500, "系统错误")
    }
    // 无数据 
}
```
