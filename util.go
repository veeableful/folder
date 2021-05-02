package folder

import (
	"math/rand"
	"strings"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Empty struct{}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func find(list []string, s string) int {
	for i, v := range list {
		if v == s {
			return i
		}
	}
	return -1
}

func remove(list []string, s string) []string {
	i := find(list, s)
	if i < 0 {
		return list
	}
	return append(list[:i], list[i+1:]...)
}

func generateRandomID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func splitWithRunes(s, runes string) (tokens []string) {
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune(runes, r) {
			return ' '
		}
		return r
	}, s)
	tokens = strings.Split(s, " ")
	return
}
