package diagnostic

type Code string

const (
	CodeUnclosedSpecialTag   Code = "UNCLOSED_SPECIAL_TAG"
	CodeUnexpectedEnd        Code = "UNEXPECTED_END"
	CodeUnclosedBlock        Code = "UNCLOSED_BLOCK"
	CodeElseWithoutIf        Code = "ELSE_WITHOUT_IF"
	CodeElseIfWithoutIf      Code = "ELSE_IF_WITHOUT_IF"
	CodeDuplicateElse        Code = "DUPLICATE_ELSE"
	CodeElseIfAfterElse      Code = "ELSE_IF_AFTER_ELSE"
	CodeInvalidRepeat        Code = "INVALID_REPEAT"
	CodeInvalidWith          Code = "INVALID_WITH"
	CodeInvalidPath          Code = "INVALID_PATH"
	CodeInvalidIndex         Code = "INVALID_INDEX"
	CodeEmptyIndex           Code = "EMPTY_INDEX"
	CodeUnclosedIndex        Code = "UNCLOSED_INDEX"
	CodeInvalidCondition     Code = "INVALID_CONDITION"
	CodeUnclosedCondition    Code = "UNCLOSED_CONDITION"
	CodeInvalidField         Code = "INVALID_FIELD"
	CodeInvalidFieldPipeline Code = "INVALID_FIELD_PIPELINE"
	CodeExpectedOperand      Code = "EXPECTED_OPERAND"
	CodeUnmatchedParenthesis Code = "UNMATCHED_PARENTHESIS"
	CodeUnknownSpecialTag    Code = "UNKNOWN_SPECIAL_TAG"
	CodeEmptyField           Code = "EMPTY_FIELD"

	CodeGoTemplateParse         Code = "GO_TEMPLATE_PARSE_ERROR"
	CodeTemplateFieldNotFound   Code = "TEMPLATE_FIELD_NOT_FOUND"
	CodeTemplateFunctionError   Code = "TEMPLATE_FUNCTION_ERROR"
	CodeTemplateSyntaxError     Code = "TEMPLATE_SYNTAX_ERROR"
	CodeTemplateExecutionFailed Code = "TEMPLATE_EXECUTION_FAILED"
	CodeInvalidFormatter        Code = "INVALID_FORMATTER"

	CodeInternalError Code = "INTERNAL_ERROR"

	CodeInvalidJSONError Code = "INVALID_JSON"
)
