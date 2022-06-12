package naml

import (
	"strings"
	"testing"
)

func TestPlain(t *testing.T) {
	type plain struct {
		Hello  string
		Number int
	}
	r := strings.NewReader(`hello = "world"  number = 123`)
	v := plain{}
	d := NewDecoder(r)
	if err := d.Decode(&v); err != nil {
		t.Fatal(err)
	}
	if v.Hello != "world" {
		t.Fatalf("wrong string: expected `world`, got `%s`", v.Hello)
	}
	if v.Number != 123 {
		t.Fatalf("wrong number: expected 123, got %d", v.Number)
	}
}

func TestBlock(t *testing.T) {
	type block struct {
		Hello string
		Names map[string]any
	}
	r := strings.NewReader(`hello = "world"  names{ a = "John" b = "Agatha" }`)
	v := block{}
	d := NewDecoder(r)
	if err := d.Decode(&v); err != nil {
		t.Fatal(err)
	}
	if v.Hello != "world" {
		t.Fatalf("wrong string: expected `world`, got `%s`", v.Hello)
	}
	if len(v.Names) < 2 {
		t.Fatalf("bad map: expected 2 keys, got %d", len(v.Names))
	}
	if _, ok := v.Names["a"]; !ok {
		t.Fatalf("bad map: could not find key `a`")
	}
	if _, ok := v.Names["b"]; !ok {
		t.Fatalf("bad map: could not find key `a`")
	}
}
