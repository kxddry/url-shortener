package validator

import (
	"regexp"
)

const (
	Fail = iota - 1
	Username
	Email
)

const (
	emailRegex = "(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])"
)

// check whether the placeholder sent by the user can be used as email or as username
// returns Email if it can be used as an email
// returns Username if it can be used as a username
// returns -1 if it can't be used as either

func ValidatePlaceholder(s string) int {
	em := ValidateEmail(s)
	if em {
		return Email
	}
	us := ValidateUsername(s)
	if us {
		return Username
	}
	return -1
}

func ValidateEmail(s string) bool {
	ss := []byte(s)
	a, _ := regexp.Match(emailRegex, ss)
	return a
}

func ValidateUsername(s string) bool {
	if len(s) < 4 || len(s) > 32 {
		return false
	}

	validChars := regexp.MustCompile(`^[a-zA-z0-9._]+$`)
	if !validChars.MatchString(s) {
		return false
	}

	return true
}

// ValidatePassword checks the password for length. The password must be in between 8 and 72 characters long.
func ValidatePassword(s string) bool {
	if len(s) < 8 || len(s) > 72 {
		return false
	}
	hasUpper := regexp.MustCompile("[A-Z]").MatchString(s)
	hasLower := regexp.MustCompile("[a-z]").MatchString(s)
	hasNumber := regexp.MustCompile("[0-9]").MatchString(s)
	hasSpecial := regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(s)

	return hasUpper && hasLower && hasNumber && hasSpecial
}
