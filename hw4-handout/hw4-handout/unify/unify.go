package unify

import (
	"errors"
    "hw4/disjointset"
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

type unifyState struct {
	sets      disjointset.DisjointSet
	termIDs   map[*term.Term]int
	desc      map[int]*term.Term
	variables []*term.Term
	nextID    int
}



// NewUnifier creates a struct of a type that satisfies the Unifier interface.
func NewUnifier() Unifier {
	return &unifier{}
}

func (u *unifier) Unify(left *term.Term, right *term.Term) (UnifyResult, error) {
	if left == nil || right == nil {
		if left == right {
			return UnifyResult{}, nil
		}
		return nil, ErrUnifier
	}

	state := newUnifyState()
	if !state.mergeTerms(left, right) {
		return nil, ErrUnifier
	}
	return state.buildResult(), nil
}

func newUnifyState() *unifyState {
	return &unifyState{
		sets:    disjointset.NewDisjointSet(),
		termIDs: make(map[*term.Term]int),
		desc:    make(map[int]*term.Term),
	}
}

func (s *unifyState) idOf(tm *term.Term) int {
	id, ok := s.termIDs[tm]
	if ok {
		return id
	}

	id = s.nextID
	s.nextID++
	s.termIDs[tm] = id
	s.sets.FindSet(id)

	if tm != nil {
		if tm.Typ == term.TermVariable {
			s.variables = append(s.variables, tm)
		} else {
			s.desc[id] = tm
			if tm.Typ == term.TermCompound {
				s.idOf(tm.Functor)
				for _, arg := range tm.Args {
					s.idOf(arg)
				}
			}
		}
	}
	return id
}

func (s *unifyState) rootOf(tm *term.Term) int {
	return s.sets.FindSet(s.idOf(tm))
}

func (s *unifyState) mergeTerms(left *term.Term, right *term.Term) bool {
	return s.mergeIDs(s.idOf(left), s.idOf(right))
}

func (s *unifyState) mergeIDs(leftID int, rightID int) bool {
	leftRoot := s.sets.FindSet(leftID)
	rightRoot := s.sets.FindSet(rightID)
	if leftRoot == rightRoot {
		return true
	}

	leftDesc := s.desc[leftRoot]
	rightDesc := s.desc[rightRoot]

	if leftDesc == nil && rightDesc != nil {
		if s.occursRoot(leftRoot, rightDesc, make(map[int]bool)) {
			return false
		}
		return s.unionClasses(leftRoot, rightRoot, rightDesc)
	}
	if rightDesc == nil && leftDesc != nil {
		if s.occursRoot(rightRoot, leftDesc, make(map[int]bool)) {
			return false
		}
		return s.unionClasses(leftRoot, rightRoot, leftDesc)
	}
	if leftDesc == nil && rightDesc == nil {
		return s.unionClasses(leftRoot, rightRoot, nil)
	}

	if leftDesc.Typ != rightDesc.Typ {
		return false
	}
	if leftDesc.Typ == term.TermAtom || leftDesc.Typ == term.TermNumber {
		if leftDesc.Literal != rightDesc.Literal {
			return false
		}
		return s.unionClasses(leftRoot, rightRoot, leftDesc)
	}
	if leftDesc.Typ != term.TermCompound {
		return false
	}
	if len(leftDesc.Args) != len(rightDesc.Args) {
		return false
	}
	if !s.mergeTerms(leftDesc.Functor, rightDesc.Functor) {
		return false
	}
	if !s.unionClasses(leftRoot, rightRoot, leftDesc) {
		return false
	}
	for i := 0; i < len(leftDesc.Args); i++ {
		if !s.mergeTerms(leftDesc.Args[i], rightDesc.Args[i]) {
			return false
		}
	}
	return true
}

func (s *unifyState) unionClasses(leftRoot int, rightRoot int, newDesc *term.Term) bool {
	newRoot := s.sets.UnionSet(leftRoot, rightRoot)
	delete(s.desc, leftRoot)
	delete(s.desc, rightRoot)
	if newDesc != nil {
		s.desc[newRoot] = newDesc
	}
	return true
}

func (s *unifyState) occursRoot(root int, tm *term.Term, seen map[int]bool) bool {
	if tm == nil {
		return false
	}
	currentRoot := s.rootOf(tm)
	if currentRoot == root {
		return true
	}

	currentDesc := s.desc[currentRoot]
	if currentDesc != nil && currentDesc != tm && !seen[currentRoot] {
		seen[currentRoot] = true
		if s.occursRoot(root, currentDesc, seen) {
			return true
		}
	}

	if tm.Typ == term.TermCompound {
		if s.occursRoot(root, tm.Functor, seen) {
			return true
		}
		for _, arg := range tm.Args {
			if s.occursRoot(root, arg, seen) {
				return true
			}
		}
	}
	return false
}

func (s *unifyState) buildResult() UnifyResult {
	result := UnifyResult{}
	classVariable := make(map[int]*term.Term)

	for _, variable := range s.variables {
		root := s.rootOf(variable)
		classDesc := s.desc[root]
		if classDesc != nil {
			result[variable] = s.rebuild(classDesc, make(map[int]bool))
		} else {
			representative, ok := classVariable[root]
			if !ok {
				classVariable[root] = variable
			} else if variable != representative {
				result[variable] = representative
			}
		}
	}
	return result
}

func (s *unifyState) rebuild(tm *term.Term, seen map[int]bool) *term.Term {
	if tm == nil {
		return nil
	}
	root := s.rootOf(tm)
	classDesc := s.desc[root]
	if classDesc != nil && classDesc != tm {
		if seen[root] {
			return tm
		}
		seen[root] = true
		return s.rebuild(classDesc, seen)
	}
	if tm.Typ != term.TermCompound {
		return tm
	}

	changed := false
	args := make([]*term.Term, len(tm.Args))
	for i, arg := range tm.Args {
		args[i] = s.rebuild(arg, seen)
		if args[i] != arg {
			changed = true
		}
	}
	if !changed {
		return tm
	}
	return &term.Term{Typ: term.TermCompound, Functor: tm.Functor, Args: args}
}