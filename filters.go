package folder

import (
	"strings"
)

var (
	punctuations = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	stopWords    = []string{"a", "and", "are", "as", "at", "be", "but", "by", "for",
		"if", "in", "into", "is", "it", "no", "not", "of", "on",
		"or", "s", "such", "t", "that", "the", "their", "then",
		"there", "these", "they", "this", "to", "was", "will",
		"with", "www"}
	stopWordsSet = MakeStringSet(stopWords)
)

// LowercaseFilter converts the tokens into their lowercase counterparts
func LowercaseFilter(tokens []string) (filteredTokens []string) {
	for _, token := range tokens {
		token = strings.ToLower(token)
		filteredTokens = append(filteredTokens, token)
	}
	return
}

// PunctuationFilter removes punctuations from tokens
func PunctuationFilter(tokens []string) (filteredTokens []string) {
	for _, token := range tokens {
		token = strings.Map(func(r rune) rune {
			if strings.ContainsRune(punctuations, r) {
				return -1
			}
			return r
		}, token)
		filteredTokens = append(filteredTokens, token)
	}
	return
}

// StopWordFilter removes tokens that are stop words
func StopWordFilter(tokens []string) (filteredTokens []string) {
	for _, token := range tokens {
		if !stopWordsSet.Contains(token) {
			filteredTokens = append(filteredTokens, token)
		}
	}
	return
}
