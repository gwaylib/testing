package database

import (
	"fmt"
	"testing"
)

func TestReflect(t *testing.T) {
	s := struct {
		// autoincrement option will be ignore to insert.
		Id int64 `db:"id,autoincrement"`
		A  int   `db:"a"`
		B  int   `db:"-"`
		C  string
	}{
		A: 100,
		B: 200,
		C: "testing",
	}

	names, inputs, vals, err := reflectInsertStruct(&s, "mysql")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", names)
	fmt.Printf("%+v\n", inputs)
	fmt.Printf("%+v\n", vals)
}
