package core

import "strings"

type LineType int

const (
	LineEmpty LineType = iota
	LineComment
	LineEnvVar
)

var sensitivePatterns = []string{
	"KEY",
	"SECRET",
	"TOKEN",
	"PASSWORD",
	"PASS",
	"AUTH",
	"PRIVATE",
	"CREDENTIAL",
}

func IsSensitiveKey(key string) bool {
	upper := strings.ToUpper(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(upper, pattern) {
			return true
		}
	}
	return false
}

type EnvLine struct {
	Type     LineType
	Key      string
	Value    string
	Raw      string
	Quote    rune
	modified bool
}

type EnvFile struct {
	Path               string
	Lines              []EnvLine
	HadTrailingNewline bool
}

func (f *EnvFile) Get(key string) (string, bool) {
	for i := range f.Lines {
		if f.Lines[i].Type == LineEnvVar && f.Lines[i].Key == key {
			return f.Lines[i].Value, true
		}
	}
	return "", false
}

func (f *EnvFile) Set(key, value string) {
	for i := range f.Lines {
		if f.Lines[i].Type == LineEnvVar && f.Lines[i].Key == key {
			f.Lines[i].Value = value
			f.Lines[i].modified = true
			return
		}
	}

	insertIdx := len(f.Lines)
	for i := len(f.Lines) - 1; i >= 0; i-- {
		if f.Lines[i].Type == LineEnvVar {
			insertIdx = i + 1
			break
		}
		if f.Lines[i].Type == LineComment {
			insertIdx = i + 1
			break
		}
	}

	newLine := EnvLine{
		Type:     LineEnvVar,
		Key:      key,
		Value:    value,
		Raw:      key + "=" + value,
		modified: true,
	}

	f.Lines = append(f.Lines[:insertIdx], append([]EnvLine{newLine}, f.Lines[insertIdx:]...)...)
}

func (f *EnvFile) Unset(key string) bool {
	for i := range f.Lines {
		if f.Lines[i].Type == LineEnvVar && f.Lines[i].Key == key {
			f.Lines = append(f.Lines[:i], f.Lines[i+1:]...)
			return true
		}
	}
	return false
}
