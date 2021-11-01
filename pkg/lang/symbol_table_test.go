package lang

import (
	"testing"
)

func Test_isDescendantOf(t *testing.T) {
	inCallable := NewType("ICall", Interface)

	human := NewType("Human", Class)
	human.implements = inCallable

	person := NewType("Person", Class)
	person.extends = human

	doctor := NewType("Doctor", Class)
	doctor.extends = person

	if !doctor.isDescendantOf(person) {
		t.Errorf("class %#v is supposed to be a descendant of %#v",
			doctor.Name(),
			person.Name(),
		)
	}

	if !doctor.isDescendantOf(human) {
		t.Errorf("class %#v is supposed to be a descendant of %#v",
			doctor.Name(),
			human.Name(),
		)
	}

	if !doctor.isImplementing(inCallable) {
		t.Errorf("class %#v is supposed to be implementing of %#v",
			doctor.Name(),
			inCallable.Name(),
		)
	}

	if !human.isImplementing(inCallable) {
		t.Errorf("class %#v is supposed to be implementing of %#v",
			human.Name(),
			inCallable.Name(),
		)
	}

}
