//go:build windows

package transformer

import "embed"

//go:embed all:weasyprint-windows
var weasyPrint embed.FS

const (
	embedDirName   = "weasyprint-windows"
	executableName = "weasyprint.exe"
)