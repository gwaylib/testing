package database

import (
	"fmt"
	"testing"
)

func TestReflect(t *testing.T) {
	s := struct {
		A int `db:"a"`
		B int
		C string
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

// TODO: api testing
