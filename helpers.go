package main

import (
	"regexp"
)

var emailRgx *regexp.Regexp = regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)

func findEmailInLine(line string) (found bool, email string) {
	lessThanIndex := -1
	greaterThanIndex := -1

	for i, char := range line {
		if char == '>' {
			greaterThanIndex = i
		} else if char == '<' && lessThanIndex == -1 {
			lessThanIndex = i
		}
	}

	if lessThanIndex == -1 ||
		greaterThanIndex == -1 ||
		lessThanIndex >= greaterThanIndex {
		return false, ""
	} else {
		email = line[lessThanIndex+1 : greaterThanIndex]
		found = emailRgx.MatchString(email)
		return
	}

}

func argSplit(str string) (args []string) {
	curr := ""
	for _, c := range str {
		if c == ' ' || c == ':' {
			if curr != "" {
				args = append(args, curr)
				curr = ""
			}
		} else {
			curr += string(c)
		}
	}
	if curr != "" {
		args = append(args, curr)
	}
	return
}
