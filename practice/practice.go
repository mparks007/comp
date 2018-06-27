package practice

import "fmt"

type person struct {
	name string
	age  int
}

// String makes person a Stringer, yes?
func (p person) String() string {
	return fmt.Sprintf("Name=%s, Age=%d", p.name, p.age)
}

// updateAgeViaValueReceiver should ignore the field update in the caller since using value receiver
func (p person) updateAgeViaValueReceiver(newAge int) {
	p.age = newAge
}

// updateAgeViaPointerReceiver should reflect the field update in the caller since using pointer receiver
func (p *person) updateAgeViaPointerReceiver(newAge int) {
	p.age = newAge
}

// updateAgeViaPersonValue should ignore the field update in the caller since using value param for person
func updateAgeViaPersonValue(p person, newAge int) {
	p.age = newAge
}

// updateAgeViaPersonPointer should reflect the field update in the caller since using pointer param for person
func updateAgeViaPersonPointer(p *person, newAge int) {
	p.age = newAge
}

// reAllocPersonViaValue should ignore the realloc of person since using value param for person
func reAllocPersonViaValue(p person) {
	p = person{}
}

// reAllocPersonViaPointer should reflect the realloc of person since using pointer param for person
func reAllocPersonViaPointer(p *person) {
	*p = person{}
}
