package database

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gwaylib/log"
	"github.com/jmoiron/sqlx/reflectx"
)

func TestReflect(t *testing.T) {
	s := struct {
		A int `db:"a"`
		B int
	}{
		A: 1,
		B: 2,
	}

	m := reflectx.NewMapper("db")
	fields := m.TypeMap(reflect.TypeOf(s))
	log.Debug(*fields.Tree)
	for i, val := range fields.Index {
		fmt.Printf("index:%d,%v\n", i, *val)
	}
	for key, val := range fields.Paths {
		fmt.Printf("path:%s,%v\n", key, *val)
	}
	for key, val := range fields.Names {
		fmt.Printf("name:%s,%v\n", key, val.Zero)
	}
}
