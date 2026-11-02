package diagnostic

type Diagnostic struct {
	Code      Code
	Stage     Stage
	Message   string
	Hint      string
	Span      *Span
	FieldPath string
}

func New(
	code Code,
	stage Stage,
	message string,
) *Diagnostic {
	result := &Diagnostic{
		Code:    code,
		Stage:   stage,
		Message: message,
	}
	return result
}

func (item *Diagnostic) WithHint(hint string) *Diagnostic {
	item.Hint = hint
	return item
}

func (item *Diagnostic) WithSpan(
	start Location,
	end Location,
) *Diagnostic {
	item.Span = &Span{
		Start: start,
		End:   end,
	}

	return item
}

func NewTranspile(
	code Code,
	message string,

) *Diagnostic {
	return New(
		code,
		StageTranspile,
		message,
	)
}
