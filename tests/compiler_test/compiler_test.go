package compiler_test

import (
	compiler "PrintServer/PrintServer/compiler"
	"PrintServer/PrintServer/diagnostic"
	"testing"
)

func TestCompilerTranspilation(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"plain text", "Привет, мир!", "Привет, мир!"},
		{"empty template", "", ""},
		{"field", "До #Field.Name# после", "До {{ .Field.Name }} после"},
		{"unicode field", "#Поле.Название#", "{{ .Поле.Название }}"},
		{"global field", "#$.Field.Something#", "{{ $.Field.Something }}"},
		{"global field inside repeat", "<#Repeat#Items#>#$.Field.Name#<#End#>", "{{ range $ItemsIndex1, $Items1 := .Items }}{{ $.Field.Name }}{{ end }}"},
		{"global indexed repeat path", "<#Repeat#$.Items[0].Children#>#Children.Name#<#End#>", "{{ range $ChildrenIndex1, $Children1 := (index $.Items 0).Children }}{{ $Children1.Name }}{{ end }}"},
		{"indexed field", "#Field[0].Name#", "{{ (index .Field 0).Name }}"},
		{"multiple indices", "#Field[0][1].Name#", "{{ (index (index .Field 0) 1).Name }}"},
		{"named index", "#Field[Selected]#", "{{ (index .Field .Selected) }}"},
		{"formatter pipeline", "#Field.Date | Weekday | Upper#", "{{ .Field.Date | Weekday | Upper }}"},
		{"formatter arguments", "#Field.Frames | FrameTimeStamp Field.FrameRate 25 \"a b\"#", "{{ .Field.Frames | FrameTimeStamp .Field.FrameRate 25 \"a b\" }}"},
		{"prefix formatter", "#Weekday Field.Date#", "{{ Weekday .Field.Date }}"},
		{"literal formatter argument", "#Format Field.Name true nil 1.5#", "{{ Format .Field.Name true nil 1.5 }}"},
		{"repeat and index", "<#Repeat#Field#>#Field.Name#:#FieldIndex#<#End#>", "{{ range $FieldIndex1, $Field1 := .Field }}{{ $Field1.Name }}:{{ $FieldIndex1 }}{{ end }}"},
		{"repeat with named index", "<#Repeat#Items#>#$.Other[ItemsIndex].Name#<#End#>", "{{ range $ItemsIndex1, $Items1 := .Items }}{{ (index $.Other $ItemsIndex1).Name }}{{ end }}"},
		{"nested repeat", "<#Repeat#Groups#><#Repeat#Items#>#Items.Name#/#Groups.Name#<#End#><#End#>", "{{ range $GroupsIndex1, $Groups1 := .Groups }}{{ range $ItemsIndex2, $Items2 := $Groups1.Items }}{{ $Items2.Name }}/{{ $Groups1.Name }}{{ end }}{{ end }}"},
		{"with alias", "<#With#Var#Field#>#Var.Name#<#End#>", "{{ with $with1 := .Field }}{{ $with1.Name }}{{ end }}"},
		{"nested with", "<#With#A#Field#><#With#B#A.Child#>#B.Name#<#End#><#End#>", "{{ with $with1 := .Field }}{{ with $with2 := $with1.Child }}{{ $with2.Name }}{{ end }}{{ end }}"},
		{"if else if else", "<#If Field.Active#>да<#ElseIf Field.Count > 0#>есть<#Else#>нет<#End#>", "{{ if .Field.Active }}да{{ else if templateGt .Field.Count 0 }}есть{{ else }}нет{{ end }}"},
		{"condition grouping", "<#If !(A == 1 || B < 2) && C#>да<#End#>", "{{ if and (not (or (templateEq .A 1) (templateLt .B 2))) .C }}да{{ end }}"},
		{"arithmetic precedence", "<#If A + B * 2 >= 5#>да<#End#>", "{{ if templateGe (add .A (mul .B 2)) 5 }}да{{ end }}"},
		{"condition string", "<#If Field.Name == 'Привет'#>да<#End#>", "{{ if templateEq .Field.Name \"Привет\" }}да{{ end }}"},
		{"condition escaped string", "<#If Name == 'a\\'b'#>да<#End#>", "{{ if templateEq .Name \"a'b\" }}да{{ end }}"},
		{"condition global indexed path", "<#If $.Items[0].Active#>да<#End#>", "{{ if (index $.Items 0).Active }}да{{ end }}"},
		{"condition consecutive not", "<#If !!Active#>да<#End#>", "{{ if not (not .Active) }}да{{ end }}"},
		{"fractional literal", "<#If Value > .5#>да<#End#>", "{{ if templateGt .Value 0.5 }}да{{ end }}"},
		{"condition scope index", "<#Repeat#Items#><#If ItemsIndex == 0#>первый<#End#><#End#>", "{{ range $ItemsIndex1, $Items1 := .Items }}{{ if templateEq $ItemsIndex1 0 }}первый{{ end }}{{ end }}"},
		{"page break", "A<#BreakPage#>B", "A<div style=\"page-break-before: always;\"></div>\nB"},
		{"escaped field", `\#Field.Name#`, `{{ "#" }}Field.Name#`},
		{"escaped special tag", `\<#End#>`, `{{ "<#End#>" }}`},
		{"escaped backslash", `\\`, `\`},
		{"ordinary backslash", `\x`, `\x`},
		{"trailing backslash", `x\`, `x\`},
		{"escaped unfinished special tag", `\<#Unknown`, `{{ "<#Unknown" }}`},
		{"literal Go template delimiters", "{{ x }}", `{{ "{{" }} x {{ "}}" }}`},
	}

	parser := compiler.NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if result.HasErrors() {
				t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
			}
			if result.Output != tt.output {
				t.Errorf("output\n got: %q\nwant: %q", result.Output, tt.output)
			}
		})
	}
}

func TestCompilerDiagnostics(t *testing.T) {
	tests := []struct {
		name, input string
		code        diagnostic.Code
	}{
		{"empty field", "##", diagnostic.CodeEmptyField},
		{"invalid field path", "#Field..Name#", diagnostic.CodeInvalidPath},
		{"global path missing field", "#$.#", diagnostic.CodeInvalidPath},
		{"invalid field after index", "#Field[0].1Name#", diagnostic.CodeInvalidPath},
		{"dangling dot", "#Field[0].#", diagnostic.CodeInvalidPath},
		{"extra closing index", "#Field]#", diagnostic.CodeInvalidPath},
		{"empty index", "#Field[]#", diagnostic.CodeEmptyIndex},
		{"invalid index", "#Field[-1]#", diagnostic.CodeInvalidIndex},
		{"index with spaces", "<#Repeat#Field[ 0]#>", diagnostic.CodeInvalidIndex},
		{"nested index", "#Field[Other[0]]#", diagnostic.CodeInvalidIndex},
		{"unclosed index", "#Field[0#", diagnostic.CodeUnclosedIndex},
		{"missing formatter", "#Field | #", diagnostic.CodeInvalidFieldPipeline},
		{"invalid formatter", "#Field | 1Wrong#", diagnostic.CodeInvalidFormatter},
		{"empty repeat", "<#Repeat##>", diagnostic.CodeInvalidRepeat},
		{"unclosed repeat tag", "<#Repeat#Field", diagnostic.CodeUnclosedSpecialTag},
		{"unclosed repeat block", "<#Repeat#Field#>", diagnostic.CodeUnclosedBlock},
		{"with missing path", "<#With#Var##>", diagnostic.CodeInvalidWith},
		{"with bad alias", "<#With#1Var#Field#>", diagnostic.CodeInvalidWith},
		{"with extra separator", "<#With#Var#Field#Other#>", diagnostic.CodeInvalidWith},
		{"with missing separator", "<#With#Var#>", diagnostic.CodeInvalidWith},
		{"unclosed with tag", "<#With#Var#Field", diagnostic.CodeUnclosedSpecialTag},
		{"unclosed with block", "<#With#Var#Field#>", diagnostic.CodeUnclosedBlock},
		{"empty if", "<#If  #>", diagnostic.CodeExpectedOperand},
		{"unclosed if tag", "<#If Field", diagnostic.CodeUnclosedSpecialTag},
		{"unclosed if block", "<#If Field#>", diagnostic.CodeUnclosedBlock},
		{"else without if", "<#Else#>", diagnostic.CodeElseWithoutIf},
		{"else if without if", "<#ElseIf Field#>", diagnostic.CodeElseIfWithoutIf},
		{"duplicate else", "<#If A#><#Else#><#Else#><#End#>", diagnostic.CodeDuplicateElse},
		{"else if after else", "<#If A#><#Else#><#ElseIf B#><#End#>", diagnostic.CodeElseIfAfterElse},
		{"else if in nested with", "<#If A#><#With#B#Field#><#ElseIf B#>", diagnostic.CodeElseIfWithoutIf},
		{"else in nested repeat", "<#If A#><#Repeat#B#><#Else#>", diagnostic.CodeElseWithoutIf},
		{"unexpected end", "<#End#>", diagnostic.CodeUnexpectedEnd},
		{"unknown special tag", "<#Unknown#>", diagnostic.CodeUnknownSpecialTag},
		{"missing right operand", "<#If A &&#>", diagnostic.CodeExpectedOperand},
		{"adjacent operands", "<#If A B#>", diagnostic.CodeExpectedOperand},
		{"unmatched left parenthesis", "<#If (A#>", diagnostic.CodeUnmatchedParenthesis},
		{"unmatched right parenthesis", "<#If A)#>", diagnostic.CodeUnmatchedParenthesis},
		{"empty parentheses", "<#If ()#>", diagnostic.CodeExpectedOperand},
		{"operator without left operand", "<#If && A#>", diagnostic.CodeExpectedOperand},
		{"not after operand", "<#If A!#>", diagnostic.CodeExpectedOperand},
		{"invalid condition character", "<#If A ? B#>", diagnostic.CodeInvalidCondition},
		{"invalid number", "<#If 12.3.4#>", diagnostic.CodeInvalidNumber},
		{"integer beyond safe range", "<#If 9007199254740993#>", diagnostic.CodeInvalidNumber},
		{"unknown string escape", "<#If A == 'a\\q'#>", diagnostic.CodeInvalidCondition},
		{"unclosed condition string", "<#If A == 'value#>", diagnostic.CodeUnclosedString},
	}

	parser := compiler.NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if !result.HasErrors() || len(result.Diagnostics) != 1 {
				t.Fatalf("expected one diagnostic, got %+v", result.Diagnostics)
			}
			got := result.Diagnostics[0]
			if got.Code != tt.code {
				t.Errorf("diagnostic code: got %s, want %s (%s)", got.Code, tt.code, got.Message)
			}
			if got.Stage != diagnostic.StageTranspile || got.Span == nil {
				t.Errorf("incomplete diagnostic: %+v", got)
			}
		})
	}
}

func TestCompilerPositionAndSessionIsolation(t *testing.T) {
	parser := compiler.NewParser()
	result := parser.Parse("first\n<#Repeat#Items#>\n#Items[]#")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %+v", result.Diagnostics)
	}
	span := result.Diagnostics[0].Span
	if span == nil || span.Start.Line != 3 || span.Start.Column != 7 {
		t.Errorf("index position: got %+v, want line 3 column 7", span)
	}
	if result.Output != "first\n{{ range $ItemsIndex1, $Items1 := .Items }}\n" {
		t.Errorf("partial output: %q", result.Output)
	}

	second := parser.Parse("<#Repeat#Items#>#Items#<#End#>")
	if second.HasErrors() || second.Output != "{{ range $ItemsIndex1, $Items1 := .Items }}{{ $Items1 }}{{ end }}" {
		t.Errorf("parser did not reset between calls: %+v", second)
	}
}
