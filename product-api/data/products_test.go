package data

import "testing"

func TestValidation(t *testing.T) {
	p := &Product{Name: "poogers", Price: 12, SKU: "ass-asss-asadsaasd"}
	err := p.Validete()

	if err != nil {
		t.Fatal(err)
	}

}
