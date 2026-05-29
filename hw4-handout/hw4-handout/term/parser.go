package term

import (
	"errors"
 // "strconv"
)

// ErrParser is the error value returned by the Parser if the string is not a
// valid term.
// See also https://golang.org/pkg/errors/#New
// and // https://golang.org/pkg/builtin/#error
var ErrParser = errors.New("parser error")

//
// <start>    ::= <term> | \epsilon
// <term>     ::= ATOM | NUM | VAR | <compound>
// <compound> ::= <functor> LPAR <args> RPAR
// <functor>  ::= ATOM
// <args>     ::= <term> | <term> COMMA <args>
//

// Parser is the interface for the term parser.
// Do not change the definition of this interface.
type Parser interface {
	Parse(string) (*Term, error)
}


type parser struct {
	lex  *lexer
	next *Token
	dag  map[string]*Term
}

// NewParser creates a struct of a type that satisfies the Parser interface.
func NewParser() Parser {
	return &parser{dag: make(map[string]*Term)}
}

func (p *parser) Parse(input string) (*Term, error) {
	p.lex = newLexer(input)
	p.next = nil

	if err := p.readNext(); err != nil {
		return nil, ErrParser
	}

	if p.next.typ == tokenEOF {
		return nil, nil
	}

	term, err := p.parseTerm()
	if err != nil {
		return nil, ErrParser
	}

	if p.next.typ != tokenEOF {
		return nil, ErrParser
	}

	return term, nil
}

func (p *parser) readNext() error {
	tok, err := p.lex.next()
	if err != nil {
		return err
	}
	p.next = tok
	return nil
}

func (p *parser) parseTerm() (*Term, error) {
	switch p.next.typ {
	case tokenAtom:
		functor := p.makeTerm(&Term{Typ: TermAtom, Literal: p.next.literal})
		if err := p.readNext(); err != nil {
			return nil, err
		}
		if p.next.typ == tokenLpar {
			return p.parseCompound(functor)
		}
		return functor, nil

	case tokenNumber:
		term := p.makeTerm(&Term{Typ: TermNumber, Literal: p.next.literal})
		if err := p.readNext(); err != nil {
			return nil, err
		}
		return term, nil

	case tokenVariable:
		term := p.makeTerm(&Term{Typ: TermVariable, Literal: p.next.literal})
		if err := p.readNext(); err != nil {
			return nil, err
		}
		return term, nil
	}

	return nil, ErrParser
}

func (p *parser) parseCompound(functor *Term) (*Term, error) {
	if err := p.readNext(); err != nil {
		return nil, err
	}

	args, err := p.parseArgs()
	if err != nil {
		return nil, err
	}

	if p.next.typ != tokenRpar {
		return nil, ErrParser
	}

	if err := p.readNext(); err != nil {
		return nil, err
	}

	return p.makeTerm(&Term{Typ: TermCompound, Functor: functor, Args: args}), nil
}

func (p *parser) parseArgs() ([]*Term, error) {
	term, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	if p.next.typ == tokenComma {
		if err := p.readNext(); err != nil {
			return nil, err
		}
		rest, err := p.parseArgs()
		if err != nil {
			return nil, err
		}
		return append([]*Term{term}, rest...), nil
	}

	return []*Term{term}, nil
}

func (p *parser) makeTerm(term *Term) *Term {
	key := termKey(term)
	oldTerm, ok := p.dag[key]
	if ok {
		return oldTerm
	}
	p.dag[key] = term
	return term
}

func termKey(term *Term) string {
	if term == nil {
		return "nil"
	}

	if term.Typ == TermAtom {
		return "atom:" + term.Literal
	}
	if term.Typ == TermNumber {
		return "number:" + term.Literal
	}
	if term.Typ == TermVariable {
		return "variable:" + term.Literal
	}

	key := "compound:" + termKey(term.Functor) + "("
	for i, arg := range term.Args {
		if i > 0 {
			key = key + ","
		}
		key = key + termKey(arg)
	}
	key = key + ")"
	return key
}