package naml

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Decoder struct {
	lex *lexer
}

func NewDecoder(src io.RuneScanner) *Decoder {
	return &Decoder{&lexer{src}}
}

func (d *Decoder) Decode(a any) error {
	m, err := d.decodeToMap()
	if err != nil {
		return err
	}

	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Pointer {
		return fmt.Errorf("can only decode to pointer values")
	}
	v = v.Elem()
	if v.Kind() == reflect.Map {
		return map2map(m, v)
	}
	if v.Kind() == reflect.Struct {
		return map2struct(m, v)
	}

	return fmt.Errorf("don't know how to decode value of type %s", v.Type())
}

func (d *Decoder) decodeToMap() (m map[string]any, err error) {
	m = map[string]any{}
	var name token
	var k string
	var v any
	for {
		name, err = d.expect(tkName)
		if errors.Is(err, io.EOF) {
			err = nil
			return
		}
		if err != nil {
			return
		}
		k, v, err = d.nextPair(name)
		if err != nil {
			return
		}
		m[k] = v
	}
}

func (d *Decoder) nextPair(nameTk token) (name string, val any, err error) {
	name = nameTk.Literal

	sep, err := d.lex.next()
	if err != nil {
		return
	}
	var k string
	var v any
	if sep.Kind == tkLBrace {
		m := map[string]any{}
		for {
			nameTk, err = d.lex.next()
			if err != nil {
				return
			}
			if nameTk.Kind == tkRBrace {
				val = m
				return
			}
			if nameTk.Kind != tkName {
				err = fmt.Errorf("expected name or right brace, got %s %s", sep.Kind, sep.Literal)
				return
			}
			k, v, err = d.nextPair(nameTk)
			if err != nil {
				return
			}
			m[k] = v
		}
	}

	var valTk token
	if sep.Kind == tkEquals {
		valTk, err = d.lex.next()
		if err != nil {
			return
		}
		switch valTk.Kind {
		case tkString:
			val = valTk.Literal
			return
		case tkNumber:
			if strings.ContainsRune(valTk.Literal, '.') {
				val, err = strconv.ParseFloat(valTk.Literal, 64)
				return
			}
			val, err = strconv.ParseInt(valTk.Literal, 0, 64)
			return
		}
		err = fmt.Errorf("expected string or number, got %s %s", valTk.Kind, valTk.Literal)
		return
	}

	err = fmt.Errorf("expected equals or left brace, got %s %s", sep.Kind, sep.Literal)
	return
}

func (d *Decoder) expect(k tokenKind) (tk token, err error) {
	tk, err = d.lex.next()
	if err != nil {
		return
	}
	if tk.Kind != k {
		err = fmt.Errorf("expected %s, got %s", k, tk.Kind)
	}
	return
}

func goName(name string) string {
	return string([]byte{name[0] &^ 0b100000}) + name[1:]
}

func map2map(src map[string]any, target reflect.Value) error {
	if target.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("can only decode to map with string keys")
	}
	if target.IsNil() {
		target.Set(reflect.MakeMap(target.Type()))
	}
	tt := target.Type().Elem()
	for k, v := range src {
		vv := reflect.ValueOf(v)
		if !vv.Type().ConvertibleTo(tt) {
			return fmt.Errorf("key %s has value of incompatible type: %s not convertible to %s", k, vv.Type(), tt)
		}
		target.SetMapIndex(reflect.ValueOf(k), vv.Convert(tt))
	}
	return nil
}

func map2struct(src map[string]any, target reflect.Value) error {
	for k, v := range src {
		f := target.FieldByName(k)
		if !f.IsValid() {
			f = target.FieldByName(goName(k))
			if !f.IsValid() {
				return fmt.Errorf("key %s has no corresponding field in %s", k, target.Type())
			}
		}
		vv := reflect.ValueOf(v)
		if vv.Kind() == reflect.Map && f.Kind() == reflect.Map {
			return map2map(v.(map[string]any), f)
		}
		if vv.Kind() == reflect.Map && f.Kind() == reflect.Struct {
			return map2struct(v.(map[string]any), f)
		}
		if !vv.Type().ConvertibleTo(f.Type()) {
			return fmt.Errorf("key %s has value of incompatible type: %s not convertible to %s", k, vv.Type(), f.Type())
		}
		f.Set(vv.Convert(f.Type()))
	}
	return nil
}
