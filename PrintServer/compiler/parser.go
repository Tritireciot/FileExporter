package template_compiler

import "PrintServer/PrintServer/diagnostic"

type blockType string

const (
	blockTypeRepeat blockType = "repeat"
	blockTypeIf     blockType = "if"
	blockTypeWith   blockType = "with"
)

type Scope struct {
	BlockType blockType
	OpenedAt  diagnostic.Location

	VarName       string
	IndexName     string
	UserIndexName string
	Fields        map[string]string

	ElseSeen bool
}

type Parser interface {
	Parse(templateStr string) ParseResult
}

type parser struct{}

func NewParser() Parser {
	return &parser{}
}
