package main

import (
	"bytes"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/russross/blackfriday/v2"
)

var (
	accountUrlRegex = regexp.MustCompile(`^(https?:\/\/)?[\w.-]+\.1password\.(com|ca|eu)\/?$`)
	urlRegex        = regexp.MustCompile(`https?://[^\s]+`)
	emailRegex      = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	emojiRegex      = regexp.MustCompile(`[\x{1F300}-\x{1F5FF}\x{1F600}-\x{1F64F}\x{1F680}-\x{1F6FF}\x{1F700}-\x{1F77F}\x{1F780}-\x{1F7FF}\x{1F800}-\x{1F8FF}\x{1F900}-\x{1F9FF}\x{1FA00}-\x{1FA6F}\x{1FA70}-\x{1FAFF}\x{1FB00}-\x{1FBFF}]+`)
	applicantRoles  = []string{"Founder or Owner", "Team Member or Employee", "Project Lead", "Core Maintainer", "Developer", "Organizer or Admin", "Program Manager"}
)

type ValidationError struct {
	Section string
	Value   string
	Message string
}

type ValidatorCallback func(string) (bool, string, string)

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Section, e.Message)
}

type Validator struct {
	Errors []ValidationError
}

func (v *Validator) AddError(section, value, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Section: section,
		Value:   value,
		Message: message,
	})
}

func (v *Validator) HasError(section string) bool {
	for _, err := range v.Errors {
		if err.Section == section {
			return true
		}
	}
	return false
}

// Parsing and validation utilities

func When(condition bool, callback ValidatorCallback) ValidatorCallback {
	if condition {
		return callback
	}

	return func(value string) (bool, string, string) {
		return true, value, ""
	}
}

func ParseInput(value string) (bool, string, string) {
	if value == "" || value == "_No response_" || value == "None" {
		return true, "", ""
	}

	return true, value, ""
}

func ParseAccountUrl(value string) (bool, string, string) {
	if accountUrlRegex.Match([]byte(value)) {
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			value = "https://" + value
		}

		u, err := url.Parse(value)
		if err != nil {
			return false, value, err.Error()
		}

		return true, u.Hostname(), ""
	} else {
		return false, value, "is an invalid 1Password account URL"
	}
}

func ParseCheckbox(value string) (bool, string, string) {
	value = strings.TrimLeft(strings.ToLower(value), "- ")

	if strings.HasPrefix(value, "[x]") {
		return true, "true", ""
	} else if strings.HasPrefix(value, "[]") || strings.HasPrefix(value, "[ ]") {
		return true, "false", ""
	}

	return false, value, "could not parse checkbox"
}

func ParseNumber(value string) (bool, int, string) {
	cleanedString := ""

	for _, char := range value {
		if char >= '0' && char <= '9' {
			cleanedString += string(char)
		}
	}

	parsedNumber, err := strconv.Atoi(cleanedString)

	if err != nil {
		return false, 0, "could not be parsed into a number"
	}

	return true, parsedNumber, ""
}

func ParseBool(value string) (bool, bool, string) {
	parsedBool, err := strconv.ParseBool(value)

	if err != nil {
		return false, false, "could not be parsed into a boolean"
	}

	return true, parsedBool, ""
}

func IsPresent(value string) (bool, string, string) {
	if value == "" {
		return false, value, "is empty"
	}

	return true, value, ""
}

func IsEmail(value string) (bool, string, string) {
	if value == "" {
		return true, value, ""
	}

	if _, err := mail.ParseAddress(value); err == nil {
		return true, value, ""
	}

	return false, value, "is an invalid email"
}

func IsUrl(value string) (bool, string, string) {
	if value == "" {
		return true, value, ""
	}

	parsedURL, err := url.ParseRequestURI(value)
	if err != nil {
		return false, value, "is an invalid URL"
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false, value, "must use \"http\" or \"https\" scheme"
	}

	return true, value, ""
}

func IsRegularString(value string) (bool, string, string) {
	// strip all formattig, except for newlines
	html := blackfriday.Run([]byte(value))
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return false, value, err.Error()
	}
	value = strings.TrimSpace(doc.Text())

	if urlRegex.MatchString(value) {
		return false, value, "cannot contain URLs"
	}

	if emailRegex.MatchString(value) {
		return false, value, "cannot contain email addresses"
	}

	if emojiRegex.MatchString(value) {
		return false, value, "cannot contain emoji characters"
	}

	return true, value, ""
}

func IsProjectRole(value string) (bool, string, string) {
	for _, item := range applicantRoles {
		if item == value {
			return true, value, ""
		}
	}

	return false, value, "is an invalid project role"
}

func IsChecked(value string) (bool, string, string) {
	if value != "true" {
		return false, value, "must be checked"
	}

	return true, value, ""
}
