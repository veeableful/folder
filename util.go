package folder

import (
	"math/rand"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Empty struct{}

func contains(list []string, s string) (yes bool) {
	for _, v := range list {
		if v == s {
			yes = true
			return
		}
	}
	return
}

func generateRandomID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
