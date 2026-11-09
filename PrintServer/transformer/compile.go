package transformer

import (
	"PrintServer/PrintServer/diagnostic"
	logging "PrintServer/agent"
	"bytes"
	"html/template"
)

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
		"add":                  add,
		"sub":                  sub,
		"mul":                  mul,
		"div":                  div,
		"templateLe":           templateLe,
		"templateEq":           templateEq,
		"templateNe":           templateNe,
		"templateGt":           templateGt,
		"templateLt":           templateLt,
		"templateGe":           templateGe,
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

func (service *TransformService) compileTemplate(content, name string) (*template.Template, error) {
	generated, err := service.parseTemplate(content)
	if err != nil {
		return nil, err
	}
	return service.prepareTemplate(generated, name)
}
