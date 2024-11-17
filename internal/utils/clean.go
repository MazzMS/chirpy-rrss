package utils

import (
	"strings"
)

func Clean(text string, badWords []string) string {
	words := strings.Fields(text)
	cleaned_words := []string{}
	for _, word := range words {
		isBadWord := false
		for _, badWord := range badWords {
			if badWord == strings.ToLower(word) {
				isBadWord = true
				break
			}
		}
		if !isBadWord {
			cleaned_words = append(cleaned_words, word)
		} else {
			cleaned_words = append(cleaned_words, "****")
		}
	}
	return strings.Join(cleaned_words, " ")
}
