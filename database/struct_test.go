package database

import (
	"fmt"
	"testing"
)

// for id insert
type ReflectTestStruct1 struct {
	Id int64 `db:"id"`
	A  int   `db:"a"`
	B  int   `db:"-"`
	C  string
}

// for autoincrement
type ReflectTestStruct2 struct {
	Id int64 `db:"id,autoincrement"`
	A  int   `db:"a"`
	B  int   `db:"-"`
	C  string
}

func (r *ReflectTestStruct2) SetLastInsertId(id int64, err error) {
	if err != nil {
		panic(err)
	}
	r.Id = id
}

func TestReflect(t *testing.T) {
	s1 := &ReflectTestStruct1{
		Id: 1,
		A:  100,
		B:  200,
		C:  "testing",
	}
	names, inputs, vals, err := reflectInsertStruct(s1, "mysql")
	if err != nil {
		t.Fatal(err)
	}
	if names != "id,a,C" {
		t.Fatal(names)
	}
	if inputs != "?,?,?" {
		t.Fatal(inputs)
	}
	if len(vals) != 3 {
		t.Fatal(fmt.Printf("%+v\n", vals))
	}

	s2 := &ReflectTestStruct2{
		A: 100,
		B: 200,
		C: "testing",
	}
	names, inputs, vals, err = reflectInsertStruct(s2, "mysql")
	if err != nil {
		t.Fatal(err)
	}
	if names != "a,C" {
		t.Fatal(names)
	}
	if inputs != "?,?" {
		t.Fatal(inputs)
	}
	if len(vals) != 2 {
		t.Fatal(fmt.Printf("%+v\n", vals))
	}
}
