package core

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func Parse(content string) (*EnvFile, error) {
	f := &EnvFile{}
	lines, err := splitLines(content)
	if err != nil {
		return nil, err
	}

	i := 0
	for i < len(lines) {
		raw := lines[i]
		line, consumed, err := parseLine(raw, lines, i)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		f.Lines = append(f.Lines, line)
		i += consumed
	}

	f.HadTrailingNewline = strings.HasSuffix(content, "\n")
	return f, nil
}

func splitLines(content string) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader([]byte(content)))
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func parseLine(raw string, all []string, idx int) (EnvLine, int, error) {
	trimmed := strings.TrimLeft(raw, " \t")

	if trimmed == "" {
		return EnvLine{Type: LineEmpty, Raw: raw}, 1, nil
	}

	if strings.HasPrefix(trimmed, "#") {
		return EnvLine{Type: LineComment, Raw: raw}, 1, nil
	}

	if strings.HasPrefix(trimmed, "export ") || strings.HasPrefix(trimmed, "export\t") {
		rest := strings.TrimSpace(trimmed[len("export"):])
		parts := strings.SplitN(rest, "=", 2)
		if len(parts) != 2 {
			return EnvLine{}, 1, fmt.Errorf("invalid export syntax: %q", raw)
		}
		key := strings.TrimRight(parts[0], " \t")
		val, quote, _ := parseValue(parts[1], all, idx)
		return EnvLine{
			Type:     LineEnvVar,
			Key:      key,
			Value:    val,
			Raw:      raw,
			Quote:    quote,
			modified: false,
		}, 1, nil
	}

	key, valueRaw, err := splitKeyValue(trimmed)
	if err != nil {
		return EnvLine{}, 1, err
	}

	value, quote, consumed := parseValue(valueRaw, all, idx)
	rawCombined := raw
	if consumed > 1 {
		rawCombined = strings.Join(all[idx:idx+consumed], "\n")
	}

	return EnvLine{
		Type:     LineEnvVar,
		Key:      key,
		Value:    value,
		Raw:      rawCombined,
		Quote:    quote,
		modified: false,
	}, consumed, nil
}

func splitKeyValue(s string) (key, value string, err error) {
	idx := strings.IndexByte(s, '=')
	if idx < 0 {
		return "", "", fmt.Errorf("missing '=' in env line: %q", s)
	}
	key = strings.TrimRight(s[:idx], " \t")
	value = s[idx+1:]
	return key, value, nil
}

func parseValue(s string, all []string, idx int) (string, rune, int) {
	trimmed := strings.TrimLeft(s, " \t")

	if trimmed == "" {
		return "", 0, 1
	}

	first := rune(trimmed[0])
	if first != '"' && first != '\'' && first != '`' {
		val := strings.TrimRight(s, " \t")
		val = stripInlineComment(val)
		return val, 0, 1
	}

	if first == '`' {
		closed, endIdx, joined := tryCollectBacktick(all, idx, trimmed)
		if closed {
			inner := joined[1 : len(joined)-1]
			return inner, '`', endIdx - idx + 1
		}
	}

	if strings.HasSuffix(trimmed, string(first)) && len(trimmed) >= 2 {
		inner := trimmed[1 : len(trimmed)-1]
		return unescape(inner, first), first, 1
	}

	val := strings.TrimRight(s, " \t")
	val = stripInlineComment(val)
	return val, 0, 1
}

func tryCollectBacktick(all []string, idx int, first string) (bool, int, string) {
	parts := []string{first}
	for i := idx + 1; i < len(all); i++ {
		parts = append(parts, all[i])
		joined := strings.Join(parts, "\n")
		if hasUnescapedClosingBacktick(joined) {
			return true, i, joined
		}
	}
	return false, idx, ""
}

func hasUnescapedClosingBacktick(s string) bool {
	if !strings.HasSuffix(s, "`") {
		return false
	}
	count := 0
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '`' {
			count++
		} else {
			break
		}
	}
	return count%2 == 1
}

func stripInlineComment(s string) string {
	idx := indexUnquoted(s, '#')
	if idx < 0 {
		return s
	}
	return strings.TrimRight(s[:idx], " \t")
}

func indexUnquoted(s string, c byte) int {
	inSingle, inDouble, inBack := false, false, false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '\'' && !inDouble && !inBack:
			inSingle = !inSingle
		case ch == '"' && !inSingle && !inBack:
			inDouble = !inDouble
		case ch == '`' && !inSingle && !inDouble:
			inBack = !inBack
		case ch == c && !inSingle && !inDouble && !inBack:
			return i
		}
	}
	return -1
}

func unescape(s string, quote rune) string {
	var b strings.Builder
	b.Grow(len(s))
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '\\' || i+1 >= len(runes) {
			b.WriteRune(r)
			continue
		}
		next := runes[i+1]
		if quote == '\'' {
			if next == '\'' {
				b.WriteRune('\'')
				i++
				continue
			}
			if next == '\\' {
				b.WriteRune('\\')
				i++
				continue
			}
			b.WriteRune(r)
			continue
		}
		if quote == '`' {
			b.WriteRune(r)
			continue
		}
		switch next {
		case 'n':
			b.WriteRune('\n')
		case 't':
			b.WriteRune('\t')
		case 'r':
			b.WriteRune('\r')
		case '\\':
			b.WriteRune('\\')
		case '"':
			b.WriteRune('"')
		default:
			b.WriteRune(r)
			b.WriteRune(next)
			i++
			continue
		}
		i++
	}
	return b.String()
}
