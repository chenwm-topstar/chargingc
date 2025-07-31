package slices

// DedupInt32s remove duplicated int32s in slice.
// Note the given slice will be modified.
func DedupInt32s(slice []int32) []int32 {
	if len(slice) <= 50 {
		return doDedupInt32sSmall(slice)
	}
	return doDedupInt32sLarge(slice)
}

// doDeupInt32Large is the hashmap version of DedupInt32 with O(n) algorithm.
func doDedupInt32sLarge(slice []int32) []int32 {
	m := map[int32]struct{}{}
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

// doDeupInt32Small is the faster version of DedupInt32 with O(n^2) algorithm.
// For n < 50, it has better performance and zero allocation.
func doDedupInt32sSmall(slice []int32) []int32 {
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

// DedupInt64s remove duplicated int64s in slice.
// Note the given slice will be modified.
func DedupInt64s(slice []int64) []int64 {
	if len(slice) <= 50 {
		return doDedupInt64sSmall(slice)
	}
	return doDedupInt64sLarge(slice)
}

// doDeupInt64Large is the hashmap version of DedupInt64 with O(n) algorithm.
func doDedupInt64sLarge(slice []int64) []int64 {
	m := map[int64]struct{}{}
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

// doDeupInt64Small is the faster version of DedupInt64 with O(n^2) algorithm.
// For n < 50, it has better performance and zero allocation.
func doDedupInt64sSmall(slice []int64) []int64 {
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

// DedupInts remove duplicated ints in slice.
// Note the given slice will be modified.
func DedupInts(slice []int) []int {
	if len(slice) <= 50 {
		return doDedupIntsSmall(slice)
	}
	return doDedupIntsLarge(slice)
}

// doDeupIntLarge is the hashmap version of DedupInt with O(n) algorithm.
func doDedupIntsLarge(slice []int) []int {
	m := map[int]struct{}{}
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

// doDeupIntSmall is the faster version of DedupInt with O(n^2) algorithm.
// For n < 50, it has better performance and zero allocation.
func doDedupIntsSmall(slice []int) []int {
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
