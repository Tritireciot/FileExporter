package template_compiler

type Scope struct {
	BlockType string            
	VarName   string            
	IndexName string
	UserIndexName string         
	Fields    map[string]string 
}

type Parser struct {
	BlockStack []Scope
	Encounter   int
}

type ParserInterface interface {
	Parse(templateStr string) string
}

func NewParser() *Parser {
	return &Parser{}
}