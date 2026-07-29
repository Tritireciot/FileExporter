package transformer

import (
	"PrintServer/html2docx/parser"
	"bytes"
	"fmt"
	"strings"

	docx "github.com/mmonterroca/docxgo/v2"
	nethtml "golang.org/x/net/html"
)

func (service *TransformService) DOCXFromTemplate(htmlContent string) ([]byte, error) {
	document := docx.NewDocument()
	rootNode, err := nethtml.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTML дерева: %v", err)
	}

	templateStyles := parser.NewTemplateStyles(rootNode)
	docParser := parser.NewDocumentParser(document, *templateStyles)
	docParser.Convert(rootNode)

	var docxBuffer bytes.Buffer
	_, err = document.WriteTo(&docxBuffer)
	if err != nil {
		return nil, fmt.Errorf("процесс конвертации в docx завершился с ошибкой: %v", err)
	}
	return docxBuffer.Bytes(), nil
}