package core

type LineType int

const (
	LineEmpty LineType = iota
	LineComment
	LineEnvVar
)

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
