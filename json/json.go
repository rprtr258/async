package json

import "fmt"

type Kind int8

const (
	KindError Kind = iota
	KindObjectBegin
	KindObjectKey
	KindObjectEnd
	KindArrayBegin
	KindArrayEnd
	KindNull
	KindTrue
	KindFalse
	KindNumber
	KindString
)

type parser struct {
	pos   int
	b     []byte
	yield func(b []byte, kind Kind)
}

func (p *parser) skip(s string) {
	for i := 0; i < len(s); i++ {
		if p.b[p.pos+i] != s[i] {
			panic(fmt.Sprintf("%s %q", string(p.b[p.pos:]), s))
		}
	}
	p.pos += len(s)
}

func (p *parser) exact(s string, kind Kind) {
	res := p.b[p.pos : p.pos+len(s)]
	p.yield(res, kind)
	p.skip(s)
}

func (p *parser) number() {
	for i := 0; ; i++ {
		switch p.b[p.pos+i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			res := p.b[p.pos:][:i]
			p.yield(res, KindNumber)
			p.pos += i
			return
		}
	}
}

func (p *parser) string(kind Kind) {
	i := 1 // skip "
	for {
		if p.b[p.pos+i] == '\\' {
			i++
		} else if p.b[p.pos+i] == '"' {
			break
		}
		i++
	}
	p.yield(p.b[p.pos:][:i+1], kind)
	p.pos += i + 1
}

func (p *parser) array() {
	p.exact("[", KindArrayBegin)
	for {
		p.json()
		if p.b[0] == ']' {
			p.exact("]", KindArrayEnd)
			break
		} else {
			p.skip(",")
		}
	}
}

func (p *parser) object() {
	p.exact("{", KindObjectBegin)
	for {
		p.string(KindObjectKey)
		p.skip(":")
		p.json()
		if p.b[p.pos] == '}' {
			p.exact("}", KindObjectEnd)
			break
		} else {
			p.skip(",")
		}
	}
}

func (p *parser) json() {
	switch p.b[p.pos] {
	case 'n':
		p.exact("null", KindNull)
	case 'f':
		p.exact("false", KindFalse)
	case 't':
		p.exact("true", KindTrue)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '-':
		p.number()
	case '{':
		p.object()
	case '[':
		p.array()
	case '"':
		p.string(KindString)
	default:
		p.yield(p.b, KindError)
		panic(fmt.Errorf("unexpected character %c", p.b[p.pos]))
	}
}

func Parse(b []byte) func(func(pos int, b []byte, kind Kind)) {
	return func(yield func(pos int, b []byte, kind Kind)) {
		p := &parser{
			pos: 0,
			b:   b,
		}
		p.yield = func(b []byte, kind Kind) {
			yield(p.pos, b, kind)
		}
		p.json()
	}
}
