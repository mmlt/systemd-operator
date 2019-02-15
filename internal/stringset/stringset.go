// Package stringset implements Set operations on a Set of strings.
package stringset

/*
Created from github.com/clipperhouse/set/templates.go by search/replace of
	{{.Pointer}} -> string
	{{.Name}} -> nothing
*/

type Set map[string]struct{}

// New creates and returns a reference to an empty set.
func New(a ...string) Set {
	s := make(Set)
	for _, i := range a {
		s.Add(i)
	}
	return s
}

// ToSlice returns the elements of the current set as a slice
func (set Set) ToSlice() []string {
	var s []string
	for v := range set {
		s = append(s, v)
	}
	return s
}

// Add adds an item to the current set if it doesn't already exist in the set.
func (set Set) Add(i string) bool {
	_, found := set[i]
	set[i] = struct{}{}
	return !found //False if it existed already
}

// Contains determines if a given item is already in the set.
func (set Set) Contains(i string) bool {
	_, found := set[i]
	return found
}

// ContainsAll determines if the given items are all in the set
func (set Set) ContainsAll(i ...string) bool {
	for _, v := range i {
		if !set.Contains(v) {
			return false
		}
	}
	return true
}

// IsSubset determines if every item in the other set is in this set.
func (set Set) IsSubset(other Set) bool {
	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

// IsSuperset determines if every item of this set is in the other set.
func (set Set) IsSuperset(other Set) bool {
	return other.IsSubset(set)
}

// Union returns a new set with all items in both sets.
func (set Set) Union(other Set) Set {
	unionedSet := New()

	for elem := range set {
		unionedSet.Add(elem)
	}
	for elem := range other {
		unionedSet.Add(elem)
	}
	return unionedSet
}

// Intersect returns a new set with items that exist only in both sets.
func (set Set) Intersect(other Set) Set {
	intersection := New()
	// loop over smaller set
	if set.Cardinality() < other.Cardinality() {
		for elem := range set {
			if other.Contains(elem) {
				intersection.Add(elem)
			}
		}
	} else {
		for elem := range other {
			if set.Contains(elem) {
				intersection.Add(elem)
			}
		}
	}
	return intersection
}

// Difference returns a new set with items in the current set but not in the other set
func (set Set) Difference(other Set) Set {
	differencedSet := New()
	for elem := range set {
		if !other.Contains(elem) {
			differencedSet.Add(elem)
		}
	}
	return differencedSet
}

// SymmetricDifference returns a new set with items in the current set or the other set but not in both.
func (set Set) SymmetricDifference(other Set) Set {
	aDiff := set.Difference(other)
	bDiff := other.Difference(set)
	return aDiff.Union(bDiff)
}

// DiffLBothR returns 3 sets; Left, Both, Right.
// Left contains the items in the current (L) set but not in the other set.
// Both contains items that are in both sets.
// Right contains the items in the other (R) set but not in the current set.
func (current Set) DiffLBothR(other Set) (Set, Set, Set) {
	l := current.Difference(other)
	both := current.Intersect(other)
	r := other.Difference(current)
	return l, both, r
}


// Clear clears the entire set to be the empty set.
func (set *Set) Clear() {
	*set = make(Set)
}

// Remove allows the removal of a single item in the set.
func (set Set) Remove(i string) {
	delete(set, i)
}

// Cardinality returns how many items are currently in the set.
func (set Set) Cardinality() int {
	return len(set)
}

// Iter returns a channel of type string that you can range over.
func (set Set) Iter() <-chan string {
	ch := make(chan string)
	go func() {
		for elem := range set {
			ch <- elem
		}
		close(ch)
	}()

	return ch
}

// Equal determines if two sets are equal to each other.
// If they both are the same size and have the same items they are considered equal.
// Order of items is not relevent for sets to be equal.
func (set Set) Equal(other Set) bool {
	if set.Cardinality() != other.Cardinality() {
		return false
	}
	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

// Clone returns a clone of the set.
// Does NOT clone the underlying elements.
func (set Set) Clone() Set {
	clonedSet := New()
	for elem := range set {
		clonedSet.Add(elem)
	}
	return clonedSet
}
