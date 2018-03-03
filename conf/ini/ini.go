package ini

import (
	"strings"
	"sync"

	"github.com/go-ini/ini"
	"github.com/gwaylib/errors"
)

var cache = sync.Map{}

func GetFile(fileName string) (*ini.File, error) {
	f, ok := cache.Load(fileName)
	if ok {
		return f.(*ini.File), nil
	}
	file, err := ini.Load(fileName)
	if err != nil {
		if strings.Index(err.Error(), "no such file or directory") > -1 {
			return nil, errors.ErrNoData.As(err, fileName)
		}
		return nil, err
	}
	cache.Store(fileName, file)
	return file, nil
}

// 用于省略前缀长路径写法
// 例如,以下可用于多语言处理：
// ini := NewIni(ini.RootDir()+"/app.default)
// lang := ".zh_cn"
// cfg := ini.Get(lang)
// cfg.Section("msg").Key("1001").String()
type Ini struct {
	rootPath string
}

func NewIni(rootPath string) *Ini {
	return &Ini{rootPath}
}

func (ini *Ini) Get(fileName string) *ini.File {
	f, err := GetFile(ini.rootPath + fileName)
	if err != nil {
		panic(err)
	}
	return f
}

func (ini *Ini) GetDefault(fileName, defFileName string) *ini.File {
	f, err := GetFile(ini.rootPath + fileName)
	if err != nil {
		if !errors.ErrNoData.Equal(err) {
			panic(err)
		}
		return ini.Get(defFileName)
	}
	return f
}
