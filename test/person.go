package test

import (
	testvar "go-ast/test/var"
)

type AgeRank int32

const (
	AgeRank_Undefined AgeRank = 0
	AgeRank_Child     AgeRank = 1
	AgeRank_Youth     AgeRank = 2
	AgeRank_MidAge    AgeRank = 3
	AgeRank_OldAge    AgeRank = 4
)

var testStr string    // 测试字符串
var testStr1 = "test" // 测试字符串
var testInt int       // 测试数字
var testInt1 = 1      // 测试数字

var testPerson Person    // 测试person
var testPersons []Person // 测试person

var testPerson2 *Person // 测试person
var testPerson3 = Person{}

var testPersons4 = []Person{}  // 测试person
var testPersons5 = []*Person{} // 测试person
var testPersons6 = &Person{}   // 测试person
// person
type Person struct {
	// id
	PersonId uint64          `json:"person_id" gorm:"primaryKey"` // id
	Name     string          `json:"name"`                        // name
	Age      uint64          `json:"age"`                         // age
	Sex      testvar.SexType `json:"sex"`                         // sex
	AgeRank  AgeRank         `json:"age_rank"`                    // age rank
}

// def
func NewDefPerson() *Person {
	return &Person{}
}

// with name
func NewPersonWithName(name string) *Person {
	return &Person{Name: name}
}

// get name
func (p *Person) GetName() string {
	return p.Name
}

// get age
func (p Person) GetAge() uint64 {
	return p.Age
}

// set name
func (p *Person) SetName(name string) {
	p.Name = name
}

// set age
func (p Person) SetAge(age uint64) {
	p.Age = age
}
