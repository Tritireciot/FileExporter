package transformer

import (
	"PrintServer/internal/db"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
)
type Transformer interface {
	RenderTemplate(ctx context.Context, template_id int, raw_data any) (string, string, error)
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

func (service *TransformService) SaveTemplate(ctx context.Context, template *db.Template) error {
	if err := service.db_repo.AddTemplate(ctx, template); err != nil {
		return err
	}
	return nil
}

func execute(template_ *template.Template, data map[string]any) (string, error) {
	var buffer bytes.Buffer
	err := template_.Execute(&buffer, data)
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

func (service *TransformService) RenderTemplate(ctx context.Context, template_id int, raw_data any) (string, string, error) {

	template_ := &db.Template{ID: template_id}
	err := service.db_repo.GetElement(ctx, template_, db.Columns.ID)

	if err != nil {
		return "", "", err
	}

	filter(template_.Subsystem, &raw_data, template_.RenderData)


	byte_data, err := json.Marshal(raw_data)
	if err != nil {
		return "", "", err
	}

	requiredTags := map[string]any{}
	repeatTags := map[string]string{}

	formatted_template := service.reshaper.TransformTemplate(ctx, template_.Content, &requiredTags, &repeatTags, template_.RenderData)

	form_template, err := template.New(template_.Name).Parse(formatted_template)

	if err != nil {
		return "", "", err
	}
	data := Translate(byte_data, requiredTags, repeatTags)
	result_template, err := execute(form_template, data)
	return template_.Name, result_template, err
}
