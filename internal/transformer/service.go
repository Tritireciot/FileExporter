package transformer

import (
	logging "PrintServer/agent"
	"PrintServer/internal/db"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
)
type Transformer interface {
	RenderTemplate(ctx context.Context, template_id int, raw_data any, subsystem string) (string, string, error)
}

type TransformService struct {
	db_repo  db.DBRepo
	reshaper Reshaper
}

func NewTransformService(db_repo db.DBRepo) *TransformService {
	return &TransformService{
		db_repo:  db_repo,
		reshaper: *NewReshaper(db_repo),
	}
}

func execute(template_ *template.Template, data map[string]any) (string, error) {
	var buffer bytes.Buffer
	err := template_.Execute(&buffer, data)
	if err != nil {
		logging.Agent.AddSimpleError("Формирование итогового шаблона для печати", "Не удалось вставить данные: " + err.Error())
	}
	return buffer.String(), err

}

func filter(subsystem string, raw_data *any, render_data map[string]bool) {
	if subsystem == "news" && raw_data != nil {
		if rundown_data, ok := (*raw_data).(map[string]any); ok {
			
			rundown, ok := rundown_data["rundown"].(map[string]any)
			if !ok {
				return
			}

			content, ok := rundown["content"].([]any)
			if !ok {
				return
			}

			filtered_content := make([]any, 0, len(content))

			for _, item := range content {
				object, ok := item.(map[string]any)
				if !ok {
					continue 
				}

				sndPrompt, _ := object["snd_prompt"].(float64)
				skipFlag, _ := object["skip_flag"].(float64)

				if render_data["snd_prompt"] && sndPrompt == 0.0 {
					fmt.Println("snd_prompt", render_data["snd_prompt"], sndPrompt)
					continue
				}
				if render_data["skip_flag"] && skipFlag == 1.0 {
					fmt.Println("skip_flag", render_data["skip_flag"], skipFlag)
					continue
				}
				
				filtered_content = append(filtered_content, object)
			}

			rundown["content"] = filtered_content
		}
	}
}

func (service *TransformService) RenderTemplate(ctx context.Context, template_id int, raw_data any, subsystem string) (string, string, error) {

	subsystem_ := &db.Subsystem{Subsystem: subsystem}
	err := service.db_repo.GetRawElement(ctx, subsystem_, db.Columns.Subsystem)

	if err != nil {
		return "", "", err
	}

	err, template_data := restRequest(ctx, subsystem_.RequestPath, raw_data)

	template_ := &db.Template{ID: template_id}
	err = service.db_repo.GetElement(ctx, template_, db.Columns.ID)
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", fmt.Sprint(template_data))

	if err != nil {
		return "", "", err
	}

	filter(template_.Subsystem, &template_data, template_.RenderData)
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Отфильтрованны данные")


	byte_data, err := json.Marshal(template_data)
	if err != nil {
		logging.Agent.AddSimpleError("Подготовка шаблона на печать", "Не удалось прочитать тело запроса: " + err.Error())
		return "", "", err
	}

	requiredTags := map[string]any{}
	repeatTags := map[string]string{}

	formatted_template := service.reshaper.TransformTemplate(ctx, template_.Content, &requiredTags, &repeatTags, template_.RenderData)
	form_template, err := template.New(template_.Name).Parse(formatted_template)

	if err != nil {
		logging.Agent.AddSimpleError("Подготовка шаблона на печать", "Не удалось сформировать шаблон: " + err.Error())
		return "", "", err
	}
	data := Translate(byte_data, requiredTags, repeatTags)
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Сформированны данные: " + fmt.Sprint(data))
	result_template, err := execute(form_template, data)
	return template_.Name, result_template, err
}
