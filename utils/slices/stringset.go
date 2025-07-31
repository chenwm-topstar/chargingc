package slices

import "errors"

// ContainString return true if given string is in the string slice.
func ContainString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// FindString return first index of given string. -1 for non existent.
func FindString(slice []string, s string) int {
	for i, item := range slice {
		if item == s {
			return i
		}
	}
	return -1
}

// OrderedSet provide fast way to find index of string in large slice.
type OrderedSet map[string]int

// NewOrderedSet return an OrderedSet object.
func NewOrderedSet(slice []string) OrderedSet {
	os := OrderedSet{}
	for i, s := range slice {
		_, hit := os[s]
		if !hit {
			os[s] = i
		}
	}
	return os
}

// Contain return true if given string is in the set.
func (os OrderedSet) Contain(s string) bool {
	_, hit := os[s]
	return hit
}

// Find return first index of given string. -1 for non existent.
func (os OrderedSet) Find(s string) int {
	idx, hit := os[s]
	if !hit {
		return -1
	}
	return idx
}

// StringSet is a data structure to determine existency.
// Note: StringSet is not thread-safe.
type StringSet map[string]struct{}

// NewStringSet returns a StringSet.
// It accepts multiple param slices to be merged all together into one set.
func NewStringSet(slices ...[]string) StringSet {
	ret := StringSet{}
	for _, slice := range slices {
		for _, s := range slice {
			ret.Add(s)
		}
	}
	return ret
}

// Contains return true if given string is in the set.
func (ss StringSet) Contains(s string) bool {
	if ss == nil {
		return false
	}
	_, hit := ss[s]
	return hit
}

// Add pushs string into set. If already in, it's a no-op.
func (ss StringSet) Add(strs ...string) error {
	if ss == nil {
		return errors.New("String Set not initialized")
	}
	for _, s := range strs {
		ss[s] = struct{}{}
	}
	return nil
}

// Merge adds all content in another string set to the current set.
func (ss StringSet) Merge(obj StringSet) error {
	if ss == nil {
		return errors.New("String Set not initialized")
	}
	for s := range obj {
		ss.Add(s)
	}
	return nil
}

// Exclude removes content in another string set from the current set.
func (ss StringSet) Exclude(obj StringSet) error {
	if ss == nil {
		return nil
	}
	for s := range obj {
		ss.Delete(s)
	}
	return nil
}

// Delete remove string from set. If not in, it's a no-op.
func (ss StringSet) Delete(strs ...string) error {
	if ss == nil {
		return nil
	}
	for _, s := range strs {
		delete(ss, s)
	}
	return nil
}

// Length of the set elements number.
func (ss StringSet) Length() int {
	if ss == nil {
		return 0
	}
	return len(ss)
}

// Slice returns a slices of elements.
func (ss StringSet) Slice() []string {
	if ss == nil {
		return nil
	}
	var ret []string
	for s := range ss {
		ret = append(ret, s)
	}
	return ret
}

// HasOverlap returns whether two StringSet have any common elements.
func (ss StringSet) HasOverlap(obj StringSet) bool {
	if ss == nil {
		return false
	}
	var a, b StringSet
	if len(ss) <= len(obj) {
		a = ss
		b = obj
	} else {
		a = obj
		b = ss
	}
	for i := range a {
		if b.Contains(i) {
			return true
		}
	}
	return false
}

// Intersect returns the intersect of two sets.
func (ss StringSet) Intersect(obj StringSet) StringSet {
	if ss == nil {
		return nil
	}
	var a, b StringSet
	if len(ss) <= len(obj) {
		a = ss
		b = obj
	} else {
		a = obj
		b = ss
	}

	ret := NewStringSet()

	for i := range a {
		if b.Contains(i) {
			ret.Add(i)
		}
	}
	return ret
}

// Union returns the union of two sets.
func (ss StringSet) Union(obj StringSet) StringSet {
	ret := NewStringSet()
	for i := range ss {
		ret.Add(i)
	}
	for i := range obj {
		ret.Add(i)
	}
	return ret
}

// Difference returns the difference of the main set exculding the objective set.
func (ss StringSet) Difference(obj StringSet) StringSet {
	if ss == nil {
		return nil
	}
	ret := NewStringSet()
	for i := range ss {
		if !obj.Contains(i) {
			ret.Add(i)
		}
	}
	return ret
}
