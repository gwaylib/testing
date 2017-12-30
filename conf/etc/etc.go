package etc

import (
	"strings"
	"sync"

	"github.com/go-ini/ini"
	"github.com/gwaylib/errors"
)

var cache = sync.Map{}

func GetEtc(fileName string) (*ini.File, error) {
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
// etc := NewEtc(etc.RootDir()+"/app.default)
// lang := ".zh_cn"
// cfg := etc.Get(lang)
// cfg.Section("msg").Key("1001").String()
type Etc struct {
	rootPath string
}

func NewEtc(rootPath string) *Etc {
	return &Etc{rootPath}
}

func (etc *Etc) Get(fileName string) *ini.File {
	f, err := GetEtc(etc.rootPath + fileName)
	if err != nil {
		panic(err)
	}
	return f
}

func (etc *Etc) GetDefault(fileName, defFileName string) *ini.File {
	f, err := GetEtc(etc.rootPath + fileName)
	if err != nil {
		if !errors.ErrNoData.Equal(err) {
			panic(err)
		}
		return etc.Get(defFileName)
	}
	return f
}
