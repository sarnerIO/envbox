package core

import (
	"fmt"
	"strings"
)

func (f *EnvFile) Render() string {
	var b strings.Builder
	for i, line := range f.Lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		switch line.Type {
		case LineEmpty, LineComment:
			b.WriteString(line.Raw)
		case LineEnvVar:
			b.WriteString(renderEnvLine(line))
		}
	}
	if f.HadTrailingNewline {
		b.WriteByte('\n')
	}
	return b.String()
}

func renderEnvLine(line EnvLine) string {
	if !line.modified {
		return line.Raw
	}
	if line.Quote == 0 {
		return fmt.Sprintf("%s=%s", line.Key, line.Value)
	}
	return fmt.Sprintf("%s=%s%s%s", line.Key, string(line.Quote), escape(line.Value, line.Quote), string(line.Quote))
}

func escape(s string, quote rune) string {
	var b strings.Builder
	b.Grow(len(s))
	if quote == '`' {
		b.WriteString(s)
		return b.String()
	}
	for _, r := range s {
		switch {
		case quote == '"' && r == '\n':
			b.WriteString(`\n`)
		case quote == '"' && r == '\t':
			b.WriteString(`\t`)
		case quote == '"' && r == '\r':
			b.WriteString(`\r`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '"' && quote == '"':
			b.WriteString(`\"`)
		case r == '\'' && quote == '\'':
			b.WriteString(`\'`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
