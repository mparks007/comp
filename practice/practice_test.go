package practice

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {

	p := person{name: "Matt",
		age: 47}

	want := fmt.Sprintf("Name=%s, Age=%d", p.name, p.age)
	if got := fmt.Sprint(p); got != want {
		t.Errorf("Wanted: %s, got: %s", want, got)
	}
}

func TestUpdateAgeViaValueReceiver(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	p.updateAgeViaValueReceiver(48)

	want := 47
	if got := p.age; got != want {
		t.Errorf("Wanted: %d, got: %d", want, got)
	}
}

func TestUpdateAgeViaPointerReceiver(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	p.updateAgeViaPointerReceiver(48)

	want := 48
	if got := p.age; got != want {
		t.Errorf("Wanted: %d, got: %d", want, got)
	}
}

func TestUpdateAgeViaPersonValue(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	updateAgeViaPersonValue(p, 48)

	want := 47
	if got := p.age; got != want {
		t.Errorf("Wanted: %d, got: %d", want, got)
	}
}

func TestUpdateAgeViaPersonPointer(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	updateAgeViaPersonPointer(&p, 48)

	want := 48
	if got := p.age; got != want {
		t.Errorf("Wanted: %d, got: %d", want, got)
	}
}

func TestReAllocPersonViaValue(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	reAllocPersonViaValue(p)

	want := 47
	if got := p.age; got != want {
		t.Errorf("Wanted: %d, got: %d", want, got)
	}
}

func TestReAllocPersonViaPointer(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	reAllocPersonViaPointer(&p)

	want := 0
	if got := p.age; got != want {
		t.Errorf("Wanted: %d, got: %d", want, got)
	}
}
