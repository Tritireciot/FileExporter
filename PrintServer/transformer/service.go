package transformer

import (
	logging "PrintServer/agent"
	"bytes"
	"fmt"
	"html/template"
)

type Transformer interface {
	RenderTemplate(template_id int, subsystem string, raw_data any) (string, string, error)
	PDFFromTemplate(htmlContent string) ([]byte, error)
	DOCXFromTemplate(htmlContent string) ([]byte, error)
}

type TransformService struct {
	reshaper Reshaper
	pdfExec  string
}

func NewTransformService() *TransformService {
	return &TransformService{
		reshaper: *NewReshaper(),
		pdfExec:  "",
	}
}

func execute(template_ *template.Template, data map[string]any) (string, error) {
	var buffer bytes.Buffer
	err := template_.Execute(&buffer, data)
	if err != nil {
		logging.Agent.AddSimpleError("Формирование итогового шаблона для печати", "Не удалось вставить данные: "+err.Error())
	}
	return buffer.String(), err

}

func (service *TransformService) RenderTemplate(template_id int, subsystem string, raw_data any) (string, string, error) {

	template_ := Template{ID: template_id, Subsystem: subsystem}
	err := getTemplate(&template_)

	if err != nil {
		logging.Agent.AddSimpleError("Подготовка шаблона на печать", "Не удалось получить шаблон: "+err.Error())
		return "", "", &TemplateNotFoundError{id: template_id}
	}

	formatted_template, err := service.reshaper.TransformTemplate(template_)
	if err != nil {
		logging.Agent.AddSimpleError("Подготовка шаблона на печать", "Не удалось сформировать шаблон: "+err.Error())
		return "", "", err
	}

	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Шаблон: "+formatted_template)
	form_template, err := template.New(template_.Name).Funcs(template.FuncMap{
		"DashedYMDtoPrintDate": DashedYMDtoPrintDate,
		"Weekday":              ShortWeekday,
		"FramesToTime":         FramesToTime,
	}).Parse(formatted_template)

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
