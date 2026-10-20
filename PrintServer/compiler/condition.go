package template_compiler

import (
	"fmt"
	"strings"
	"unicode"
)

type Node interface {
	String() string
}

type LiteralNode struct {
	Value string
	Parser *Parser
}

func (n *LiteralNode) String() string {

	if len(n.Value) == 0 {
		return ""
	}
	
	if strings.HasPrefix(n.Value, `"`) || strings.HasPrefix(n.Value, `'`) {
		return n.Value
	}
	if unicode.IsDigit(rune(n.Value[0])) || n.Value == "true" || n.Value == "false" {
		return n.Value
	}
	

	if strings.Contains(n.Value, "[") && strings.Contains(n.Value, "]") {
		if n.Parser != nil {
			return n.Parser.convertIndexes(n.Value)
		}
	}
	
	if n.Parser != nil {
		replaced := n.Parser.ReplaceVar(n.Value)
		if len(replaced) > 0 && replaced[0] == '$' {
			return replaced
		}
		return "." + replaced
	}
	
	return "." + n.Value
}

// Узел для оператора (например, "==", "&&")
type BinaryNode struct {
	Op    string
	Left  Node
	Right Node
}

// Карта соответствия операторов Go-шаблонам
var templateOps = map[string]string{
	"||": "or",
	"&&": "and",
	"==": "eq",
	"!=": "ne",
	"<":  "lt",
	">":  "gt",
	"<=": "le",
	">=": "ge",
	"+":  "add",
	"-":  "sub",
	"*":  "mul",
	"/":  "div",
}

func (n *BinaryNode) String() string {
	goOp := templateOps[n.Op]
	if goOp == "" {
		goOp = n.Op // на случай +, -, *, /
	}
	// Синтаксис Go-шаблонов: (оператор левый_аргумент правый_аргумент)
	return fmt.Sprintf("(%s %s %s)", goOp, n.Left.String(), n.Right.String())
}

type NotNode struct {
	Child Node
}

func (n *NotNode) String() string {
	// Переводим в префиксный синтаксис Go-шаблонов: (not выражение)
	return fmt.Sprintf("(not %s)", n.Child.String())
}

// --- Парсер на основе алгоритма Сортировочной станции ---

func opPrecedence(op string) int {
	switch op {
	case "||":
		return 1
	case "&&":
		return 2
	case "==", "!=", "<", ">", "<=", ">=":
		return 3
	case "+", "-":
		return 4
	case "*", "/":
		return 5
	case "!": 
		return 6
	}
	return 0
}

func tokenize(expr string) []string {
	var tokens []string
	runes := []rune(expr)
	i := 0
	for i < len(runes) {
		if unicode.IsSpace(runes[i]) {
			i++
			continue
		}

		if unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '$' || runes[i] == '.' || runes[i] == '_' {
			start := i
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '$' || runes[i] == '.' || runes[i] == '_' || runes[i] == '[' || runes[i] == ']') {
				i++
			}
			token := string(runes[start:i])
			tokens = append(tokens, token)
			continue
		}

		if runes[i] == '"' || runes[i] == '\'' {
			quote := runes[i]
			start := i
			i++ 
			for i < len(runes) && runes[i] != quote {
				if runes[i] == '\\' && i+1 < len(runes) {
					i += 2
					continue
				}
				i++
			}
			if i < len(runes) { i++ } 
			tokens = append(tokens, string(runes[start:i]))
			continue
		}

		if runes[i] == '(' || runes[i] == ')' {
			tokens = append(tokens, string(runes[i]))
			i++
			continue
		}
		if i+1 < len(runes) {
			nextTwo := string(runes[i : i+2])
			if _, ok := templateOps[nextTwo]; ok {
				tokens = append(tokens, nextTwo)
				i += 2
				continue
			}
		}
		if strings.ContainsRune("+-*/><!", runes[i]) {
			tokens = append(tokens, string(runes[i]))
			i++
			continue
		}
		if unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '$' ||
			runes[i] == '.' ||
			runes[i] == '_' {
			start := i
			for i < len(runes) && (unicode.IsLetter(runes[i]) || runes[i] == '$' || runes[i] == '.' || runes[i] == '_' || unicode.IsDigit(runes[i]) || runes[i] == '.') {
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}
		i++
	}
	return tokens
}

func (parser *Parser) ReplaceVar(token string) string {
	if len(token) == 0 {
		return token
	}

	if token[0] == '"' || token[0] == '\'' {
		return token
	}

	for i := len(parser.BlockStack) - 1; i >= 0; i-- {
		scope := parser.BlockStack[i]

		if scope.BlockType == "if" {
			continue
		}
		
		if scope.UserIndexName != "" && token == scope.UserIndexName {
			return scope.IndexName
		}
		
		for key, varName := range scope.Fields {
			if token == key {
				return varName
			}
			if strings.HasPrefix(token, key+".") {
				return varName + token[len(key):]
			}
		}
	}
	return token
}

// Строит дерево выражений
func (parser *Parser) parseToAST(expr string) Node {
	tokens := tokenize(expr)
	var nodeStack []Node
	var opStack []string

	buildNode := func() {

		if len(opStack) == 0 { return }
		op := opStack[len(opStack)-1]
		opStack = opStack[:len(opStack)-1]

		if op == "!" {
			if len(nodeStack) < 1 { return }
			child := nodeStack[len(nodeStack)-1]
			nodeStack = nodeStack[:len(nodeStack)-1]
			
			nodeStack = append(nodeStack, &NotNode{Child: child})
			return
		}

		if len(nodeStack) < 2 { return }
		right := nodeStack[len(nodeStack)-1]
		left := nodeStack[len(nodeStack)-2]
		nodeStack = nodeStack[:len(nodeStack)-2]

		nodeStack = append(nodeStack, &BinaryNode{Op: op, Left: left, Right: right})
	}

	for _, token := range tokens {
		if token == "(" {
			opStack = append(opStack, token)
		} else if token == ")" {
			for len(opStack) > 0 && opStack[len(opStack)-1] != "(" {
				buildNode()
			}
			if len(opStack) > 0 {
				opStack = opStack[:len(opStack)-1]
			} 
		} else if prec := opPrecedence(token); prec > 0 {
			if token == "!" {
				for len(opStack) > 0 && opPrecedence(opStack[len(opStack)-1]) > prec {
					buildNode()
				}
			} else {
				for len(opStack) > 0 && opPrecedence(opStack[len(opStack)-1]) >= prec {
					buildNode()
				}
			}
			opStack = append(opStack, token)
		} else {
			nodeStack = append(nodeStack, &LiteralNode{Value: parser.ReplaceVar(token), Parser: parser})
		}
	}

	for len(opStack) > 0 {
		buildNode()
	}

	if len(nodeStack) == 0 {
		return nil
	}
	return nodeStack[0]
}

func (parser *Parser) parseCondition(condition string) string {
	if strings.HasPrefix(condition, "Condition:") {
		condition = condition[len("Condition:")+1:]
	}
	ast := parser.parseToAST(condition)
	if ast == nil {
		return ""
	}

	res := ast.String()
	if strings.HasPrefix(res, "(") && strings.HasSuffix(res, ")") {
		res = res[1 : len(res)-1]
	}
	return res
}
