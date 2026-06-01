package auth

import (
	"slices"
	"strconv"
)

func toIntSet(s []int64) map[int64]bool {
	m := make(map[int64]bool, len(s))
	for _, v := range s {
		m[v] = true
	}
	return m
}

func toNameSet(s []string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, v := range s {
		m[normalize(v)] = true
	}
	return m
}

func removeInt(s []int64, v int64) []int64 {
	return slices.DeleteFunc(s, func(x int64) bool { return x == v })
}

func removeStr(s []string, v string) []string {
	return slices.DeleteFunc(s, func(x string) bool { return x == v })
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
