package disjointset

// DisjointSet is the interface for the disjoint-set (or union-find) data
// structure.
// Do not change the definition of this interface.
type DisjointSet interface {
	// UnionSet(s, t) merges (unions) the sets containing s and t,
	// and returns the representative of the resulting merged set.
	UnionSet(int, int) int
	// FindSet(s) returns representative of the class that s belongs to.
	FindSet(int) int
}

// TODO: implement a type that satisfies the DisjointSet interface.

type disjointSet struct {
	parent map[int]int
	rank   map[int]int
}


// NewDisjointSet creates a struct of a type that satisfies the DisjointSet interface.
func NewDisjointSet() DisjointSet {
	return &disjointSet{
		parent: make(map[int]int),
		rank:   make(map[int]int),
	}
}

func (d *disjointSet) FindSet(x int) int {
	parent, ok := d.parent[x]
	if !ok {
		d.parent[x] = x
		d.rank[x] = 0
		return x
	}

	if parent != x {
		d.parent[x] = d.FindSet(parent)
	}

	return d.parent[x]
}

func (d *disjointSet) UnionSet(x int, y int) int {
	rootX := d.FindSet(x)
	rootY := d.FindSet(y)

	if rootX == rootY {
		return rootX
	}

	if d.rank[rootX] < d.rank[rootY] {
		d.parent[rootX] = rootY
		return rootY
	}

	d.parent[rootY] = rootX

	if d.rank[rootX] == d.rank[rootY] {
		d.rank[rootX] = d.rank[rootX] + 1
	}

	return rootX
}