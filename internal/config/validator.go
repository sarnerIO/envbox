package config

import (
	"net/url"
	"strconv"
	"strings"
)

type ValidationResult struct {
	Key    string
	Type   string
	Valid  bool
	Reason string
}

func Validate(key, value, fieldType string) ValidationResult {
	switch fieldType {
	case "string":
		return validateString(key, value)
	case "number":
		return validateNumber(key, value)
	case "url":
		return validateURL(key, value)
	case "boolean":
		return validateBoolean(key, value)
	default:
		return ValidationResult{Key: key, Type: fieldType, Valid: false, Reason: "unknown type"}
	}
}

func validateString(key, value string) ValidationResult {
	return ValidationResult{Key: key, Type: "string", Valid: true}
}

func validateNumber(key, value string) ValidationResult {
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		return ValidationResult{Key: key, Type: "number", Valid: false, Reason: "not a number"}
	}
	return ValidationResult{Key: key, Type: "number", Valid: true}
}

func validateURL(key, value string) ValidationResult {
	u, err := url.Parse(value)
	if err != nil {
		return ValidationResult{Key: key, Type: "url", Valid: false, Reason: "invalid url"}
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ValidationResult{Key: key, Type: "url", Valid: false, Reason: "url must start with http:// or https://"}
	}
	return ValidationResult{Key: key, Type: "url", Valid: true}
}

func validateBoolean(key, value string) ValidationResult {
	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" || lower == "1" || lower == "0" {
		return ValidationResult{Key: key, Type: "boolean", Valid: true}
	}
	return ValidationResult{Key: key, Type: "boolean", Valid: false, Reason: "must be true/false"}
}
