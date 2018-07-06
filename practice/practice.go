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

// updateAgeViaPointerToPointerReceiver should reflect the pointing of the parameter to a different person struct
// func (p **person) updateAgeViaPointerToPointerReceiver(newAge int) {
// 	p2 := &person{"Bob", 50}
// 	*p = p2 // one dereference level to get to the pointer of the parameter, then make it point elsewhere (to p2)
// }

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

// reAllocPersonViaPointerToPointer should reflect the pointing of the parameter to a different person struct
func reAllocPersonViaPointerToPointer(p **person) {
	p2 := &person{"Bob", 50}
	*p = p2 // one dereference level to get to the pointer of the parameter, then make it point elsewhere (to p2)
}
