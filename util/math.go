package util

func Max(i ...int) int {
	if len(i) == 0 {
		return 0
	}
	m := i[0]
	for _, e := range i[1:] {
		if e > m {
			m = e
		}
	}
	return m
}
