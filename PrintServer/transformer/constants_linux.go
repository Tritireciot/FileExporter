//go:build linux

package transformer

import "embed"

//go:embed all:weasyprint-linux
var weasyPrint embed.FS

const (
	embedDirName   = "weasyprint-linux"
	executableName = "weasyprint"
)