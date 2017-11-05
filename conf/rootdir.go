package conf

import (
	"os"
)

func RootDir() string {
	dir := os.Getenv("PJ_ROOT")
	if len(dir) == 0 {
		panic("Need PJ_ROOT environment for project directory")
	}

	if dir[len(dir)-1:] == "/" {
		return dir[len(dir)-1:]
	}
	return dir
}
