package slices

func Reverse(slice []string) []string {
	ret := Clone(slice)
	for i := len(ret)/2 - 1; i >= 0; i-- {
		opp := len(ret) - 1 - i
		ret[i], ret[opp] = ret[opp], ret[i]
	}
	return ret
}
