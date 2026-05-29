package unify

import (
	"errors"
 // "hw4/disjointset"
	"hw4/term"
)

// ErrUnifier is the error value returned by the Parser if the string is not a
// valid term.
// See also https://golang.org/pkg/errors/#New
// and // https://golang.org/pkg/builtin/#error
var ErrUnifier = errors.New("unifier error")

// UnifyResult is the result of unification. For example, for a variable term
// `s`, `UnifyResult[s]` is the term which `s` is unified with.
type UnifyResult map[*term.Term]*term.Term

// Unifier is the interface for the term unifier.
// Do not change the definition of this interface
type Unifier interface {
	Unify(*term.Term, *term.Term) (UnifyResult, error)


}


type unifier struct{}


// NewUnifier creates a struct of a type that satisfies the Unifier interface.
func NewUnifier() Unifier {
	return &unifier{}
}

func (u *unifier) Unify(left *term.Term, right *term.Term) (UnifyResult, error) {
	result := UnifyResult{}
	if !unifyTerms(left, right, result) {
		return nil, ErrUnifier
	}

	for variable, value := range result {
		result[variable] = substitute(value, result)
	}
	return result, nil
}

func unifyTerms(left *term.Term, right *term.Term, result UnifyResult) bool {
	left = substitute(left, result)
	right = substitute(right, result)

	if left == right {
		return true
	}
	if left == nil || right == nil {
		return left == right
	}
	if left.Typ == term.TermVariable {
		return bind(left, right, result)
	}
	if right.Typ == term.TermVariable {
		return bind(right, left, result)
	}
	if left.Typ != right.Typ {
		return false
	}
	if left.Typ == term.TermAtom || left.Typ == term.TermNumber {
		return left.Literal == right.Literal
	}

	if !unifyTerms(left.Functor, right.Functor, result) {
		return false
	}
	if len(left.Args) != len(right.Args) {
		return false
	}
	for i := 0; i < len(left.Args); i++ {
		if !unifyTerms(left.Args[i], right.Args[i], result) {
			return false
		}
	}
	return true
}

func bind(variable *term.Term, value *term.Term, result UnifyResult) bool {
	value = substitute(value, result)
	if variable == value {
		return true
	}
	if occurs(variable, value, result) {
		return false
	}
	result[variable] = value
	return true
}

func substitute(current *term.Term, result UnifyResult) *term.Term {
	if current == nil {
		return nil
	}
	if current.Typ == term.TermVariable {
		value, ok := result[current]
		if ok {
			return substitute(value, result)
		}
		return current
	}
	if current.Typ != term.TermCompound {
		return current
	}

	changed := false
	newArgs := make([]*term.Term, len(current.Args))
	for i := 0; i < len(current.Args); i++ {
		newArgs[i] = substitute(current.Args[i], result)
		if newArgs[i] != current.Args[i] {
			changed = true
		}
	}
	if !changed {
		return current
	}
	return &term.Term{Typ: term.TermCompound, Functor: current.Functor, Args: newArgs}
}

func occurs(variable *term.Term, current *term.Term, result UnifyResult) bool {
	current = substitute(current, result)
	if current == nil {
		return false
	}
	if current == variable {
		return true
	}
	if current.Typ != term.TermCompound {
		return false
	}
	for i := 0; i < len(current.Args); i++ {
		if occurs(variable, current.Args[i], result) {
			return true
		}
	}
	return false
}