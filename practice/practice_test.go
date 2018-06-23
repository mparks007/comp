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
		t.Errorf("Wanted:%s, got:%s", want, got)
	}
}

func TestAgeUpdate_Direct(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	p.age = 48

	want := 48
	if got := p.age; got != want {
		t.Errorf("Wanted:%d, got:%d", want, got)
	}
}

func TestUpdateAgeAsValueReceiver(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	p.updateAgeAsValueReceiver(48)

	want := 47
	if got := p.age; got != want {
		t.Errorf("Wanted:%d, got:%d", want, got)
	}
}

func TestUpdateAgeAsPointerReceiver(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	p.updateAgeAsPointerReceiver(48)

	want := 48
	if got := p.age; got != want {
		t.Errorf("Wanted:%d, got:%d", want, got)
	}
}

func TestUpdateAgeAsPersonValue(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	updateAgeAsPersonValue(p, 48)

	want := 47
	if got := p.age; got != want {
		t.Errorf("Wanted:%d, got:%d", want, got)
	}
}

func TestUpdateAgeAsPersonPointer(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	updateAgeAsPersonPointer(&p, 48)

	want := 48
	if got := p.age; got != want {
		t.Errorf("Wanted:%d, got:%d", want, got)
	}
}

func TestReAllocPersonPointer(t *testing.T) {
	p := person{name: "Matt",
		age: 47}

	reAllocPersonPointer(&p, 48)

	want := 0
	if got := p.age; got != want {
		t.Errorf("Wanted:%d, got:%d", want, got)
	}
}
