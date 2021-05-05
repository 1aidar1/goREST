package data

import "testing"

func TestValidation(t *testing.T) {
	p := &Book{}
	err := p.Validate()

	if err != nil {
		t.Fatal(err)
	}

}
