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

type ReflectTestStruct3 struct {
	ReflectTestStruct1
	D string `db:"d"`
}
type ReflectTestStruct4 struct {
	ReflectTestStruct3
	*ReflectTestStruct2
	E string `db:"e"`
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
		A: 101,
		B: 201,
		C: "testing1",
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

	s4 := &ReflectTestStruct4{
		ReflectTestStruct2: s2,
		ReflectTestStruct3: ReflectTestStruct3{
			ReflectTestStruct1: *s1,
			D:                  "d",
		},
		E: "e",
	}
	names, inputs, vals, err = reflectInsertStruct(s4, "oracle")
	if err != nil {
		t.Fatal(err)
	}
	if string(names) != "id,a,C,d,a,C,e" {
		t.Fatal(string(names))
	}
	if string(inputs) != ":id,:a,:C,:d,:a,:C,:e" {
		t.Fatal(string(inputs))
	}
	if fmt.Sprintf("%+v", vals) != "[1 100 testing d 101 testing1 e]" {
		t.Fatal(vals)
	}
}
