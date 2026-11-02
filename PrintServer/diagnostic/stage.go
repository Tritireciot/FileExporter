package diagnostic

type Stage string

const (
	StageTranspile       Stage = "transpile"
	StageGoTemplateParse Stage = "parsing"
	StageExecute         Stage = "execution"
	StageInternal        Stage = "internal"
	StageRequest        Stage = "request"
)
