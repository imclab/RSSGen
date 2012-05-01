package levenshtein

func Min(a ...int) int {
	min := int(^uint(0) >> 1)
	for _, i := range a {
		if i < min { min = i }
	}
	return min
}

func Levenshtein(s1, s2 string) int {
	if len(s1) < len(s2) {
		return Levenshtein(s2, s1)
	}
	
	prev := make([]int, len(s2) + 1)
	for k, _ := range prev { prev[k] = k }
	
	for i, c1 := range s1 {
		cur := []int{i + 1}
		for j, c2 := range s2 {
			ins := prev[j + 1] + 1
			del := cur[j] + 1
			sub := prev[j]
			if c1 != c2 { sub++ }
			cur = append(cur, Min(ins, del, sub))
		}
		prev = cur
	}
	return prev[len(prev) - 1]
}
