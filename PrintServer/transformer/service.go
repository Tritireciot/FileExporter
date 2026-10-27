package transformer

import (
	"PrintServer/PrintServer/compiler"
	"PrintServer/PrintServer/diagnostic"
	logging "PrintServer/agent"
	"bytes"
	"fmt"
	"html/template"
)

type Transformer interface {
	RenderTemplate(templateID int, subsystem string, rawData any) (string, string, error)
	PreviewTemplate(templateContent string, subsystem string, templateType string) (string, string, error)
	ValidateTemplate(templateContent string, subsystem string, templateType string, contentHash string, templateData map[string]any) error
	PDFFromTemplate(htmlContent string) ([]byte, error)
	DOCXFromTemplate(htmlContent string) ([]byte, error)
}

type TransformService struct {
	parser  template_compiler.Parser
	pdfExec string
}

func NewTransformService() *TransformService {
	return &TransformService{
		parser:  template_compiler.NewParser(),
		pdfExec: "",
	}
}

func execute(template_ *template.Template, data map[string]any) (string, error) {
	var buffer bytes.Buffer
	err := template_.Execute(&buffer, data)
	if err != nil {
		logging.Agent.AddSimpleError("Формирование итогового шаблона для печати", "Не удалось вставить данные: "+err.Error())
		return "", &diagnostic.TemplateError{
			Diagnostics: []diagnostic.Diagnostic{
				{
					Code:    diagnostic.CodeTemplateExecutionFailed,
					Stage:   diagnostic.StageExecute,
					Message: "Не удалось выполнить шаблон",
				},
			},
			Cause: err,
		}
	}
	return buffer.String(), nil

}

func (service *TransformService) prepareTemplate(formattedTemplate string, templateName string) (*template.Template, error) {
	form_template, err := template.New(templateName).Funcs(template.FuncMap{
		"DashedYMDtoPrintDate": DashedYMDtoPrintDate,
		"Weekday":              ShortWeekday,
		"FramesToTime":         FramesToTime,
	}).Option("missingkey=error").
		Parse(formattedTemplate)

	if err != nil {
		return nil, &diagnostic.TemplateError{
			Diagnostics: []diagnostic.Diagnostic{
				{
					Code:    diagnostic.CodeGoTemplateParse,
					Stage:   diagnostic.StageGoTemplateParse,
					Message: "Не удалось скомпилировать шаблон",
				},
			},
			Cause: err,
		}
	}
	return form_template, nil
}

func (service *TransformService) parseTemplate(templateContent string) (string, error) {
	parseResult := service.parser.Parse(templateContent)
	if parseResult.HasErrors() {
		return "", &diagnostic.TemplateError{
			Diagnostics: parseResult.Diagnostics,
		}
	}
	return parseResult.Output, nil
}

func (service *TransformService) RenderTemplate(template_id int, subsystem string, raw_data any) (string, string, error) {

	template_ := Template{ID: template_id, Subsystem: subsystem}
	err := getTemplate(&template_)

	if err != nil {
		logging.Agent.AddSimpleError("Подготовка шаблона на печать", "Не удалось получить шаблон: "+err.Error())
		return "", "", &TemplateNotFoundError{id: template_id}
	}

	generatedTemplate, err := service.parseTemplate(template_.Content)
	if err != nil {
		return "", "", err
	}
	form_template, err := service.prepareTemplate(generatedTemplate, template_.Name)

	if err != nil {
		logging.Agent.AddSimpleError("Подготовка шаблона на печать", "Не удалось сформировать шаблон: "+err.Error())
		return "", "", err
	}
	data, err := getPrintData(raw_data, &template_)
	if err != nil {
		logging.Agent.AddSimpleError("Подготовка шаблона печати", err.Error())
		return "", "", err
	}
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Сформированны данные: "+fmt.Sprint(data))
	result_template, err := execute(form_template, data)
	return template_.Name, result_template, err
}

func (service *TransformService) PreviewTemplate(templateContent string, subsystem string, templateType string) (string, string, error) {
	generatedTemplate, err := service.parseTemplate(templateContent)
	if err != nil {
		return "", "", err
	}
	form_template, err := service.prepareTemplate(generatedTemplate, templateType)
	if err != nil {
		logging.Agent.AddSimpleError("Предпросмотр шаблона", "Не удалось сформировать шаблон: "+err.Error())
		return "", "", err
	}
	data, err := getPreviewData(subsystem, templateType)
	if err != nil {
		logging.Agent.AddSimpleError("Предпросмотр шаблон", err.Error())
		return "", "", err
	}
	result_template, err := execute(form_template, data)
	return result_template, templateType, err
}

func (service *TransformService) ValidateTemplate(
	templateContent string,
	subsystem string,
	templateType string,
	contentHash string,
	templateData map[string]any,
) error {
	generatedTemplate, err := service.parseTemplate(templateContent)
	if err != nil {
		return err
	}
	form_template, err := service.prepareTemplate(generatedTemplate, templateType)
	if err != nil {
		return err
	}

	_, err = execute(form_template, templateData)

	if err != nil {
		return err
	}

	return nil
}
