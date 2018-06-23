package practice

import "fmt"

type person struct {
	name string
	age  int
}

func (p person) String() string {
	return fmt.Sprintf("Name=%s, Age=%d", p.name, p.age)
}

func (p person) updateAgeAsValueReceiver(newAge int) {
	p.age = newAge
}

func (p *person) updateAgeAsPointerReceiver(newAge int) {
	*p = person{}
	//	p.age = newAge
}

func updateAgeAsPersonValue(p person, newAge int) {
	p.age = newAge
}

func updateAgeAsPersonPointer(p *person, newAge int) {
	p.age = newAge
}

func reAllocPersonValue(p person, newAge int) {
	p = person{}
}

func reAllocPersonPointer(p *person, newAge int) {
	*p = person{}
}
