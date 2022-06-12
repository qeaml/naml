package naml

import (
	"errors"
	"io"
)

type token struct {
	Kind    tokenKind
	Literal string
}

type lexer struct {
	r io.RuneScanner
}

func (l *lexer) next() (token, error) {
	r, _, err := l.r.ReadRune()
	if err != nil {
		return token{tkInvalid, ""}, err
	}

	for isWhitespace(r) {
		r, _, err = l.r.ReadRune()
		if err != nil {
			return token{tkInvalid, ""}, err
		}
	}

	if isDigit(r) || r == '-' {
		return l.nextNumber(r)
	}

	switch r {
	case '=':
		return token{tkEquals, "="}, nil
	case '{':
		return token{tkLBrace, "{"}, nil
	case '}':
		return token{tkRBrace, "}"}, nil
	case '"':
		return l.nextString()
	}

	// spec does not specify what is and what's not a valid name, so we'll just
	// assume anything goes :)
	return l.nextName(r)
}

func (l *lexer) nextString() (token, error) {
	literal := ""
	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			return token{tkInvalid, ""}, err
		}
		if r == '"' {
			break
		}
		literal += string(r)
	}
	return token{tkString, literal}, nil
}

func (l *lexer) nextNumber(r rune) (token, error) {
	literal := string(r)
	for {
		r, _, err := l.r.ReadRune()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return token{tkInvalid, ""}, err
		}
		if !isDigit(r) && r != '.' && r != '_' {
			if err := l.r.UnreadRune(); err != nil {
				return token{tkInvalid, ""}, err
			}
			break
		}
		literal += string(r)
	}
	return token{tkNumber, literal}, nil
}

func (l *lexer) nextName(r rune) (token, error) {
	name := string(r)
	for {
		r, _, err := l.r.ReadRune()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return token{tkInvalid, ""}, err
		}
		if r == '=' || r == '{' || r == '}' || isWhitespace(r) || isDigit(r) || r == '-' {
			if err := l.r.UnreadRune(); err != nil {
				return token{tkInvalid, ""}, err
			}
			break
		}
		name += string(r)
	}
	return token{tkName, name}, nil
}

type tokenKind uint8

const (
	tkInvalid tokenKind = iota
	tkName
	tkEquals
	tkLBrace
	tkRBrace
	tkString
	tkNumber
)

var kindNames = []string{
	"invalid",
	"name",
	"equals",
	"left brace",
	"right brace",
	"string",
	"number",
}

func (k tokenKind) String() string {
	return kindNames[k]
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isWhitespace(r rune) bool {
	return r == ' ' ||
		r == '\t' ||
		r == '\n' ||
		r == '\f' ||
		r == '\v' ||
		r == ';' // lol
}
