package slices

// DedupStrings remove duplicated strings in slice.
// Note the given slice will be modified.
func DedupStrings(slice []string) []string {
	if len(slice) <= 50 {
		return doDedupStringsSmall(slice)
	}
	return doDedupStringsLarge(slice)
}

// doDeupStringLarge is the hashmap version of DedupString with O(n) algorithm.
func doDedupStringsLarge(slice []string) []string {
	m := map[string]struct{}{}
	idx := 0
	for i, s := range slice {
		if _, hit := m[s]; hit {
			continue
		} else {
			m[s] = struct{}{}
			slice[idx] = slice[i]
			idx++
		}
	}

	return slice[:idx]
}

// doDeupStringSmall is the faster version of DedupString with O(n^2) algorithm.
// For n < 50, it has better performance and zero allocation.
func doDedupStringsSmall(slice []string) []string {
	idx := 0
	for _, s := range slice {
		var j int
		for j = 0; j < idx; j++ {
			if slice[j] == s {
				break
			}
		}
		if j >= idx {
			slice[idx] = s
			idx++
		}
	}

	return slice[:idx]
}
