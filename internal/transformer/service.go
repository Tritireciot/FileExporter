package transformer

import (
	"bytes"
	"context"
	"former/internal/db"
	"html/template"
)

type TransformService struct {
	db_repo        *db.DBRepository
	reshaper       Reshaper
}

func NewTransformService(db_repo *db.DBRepository) *TransformService {
	return &TransformService{
		db_repo:        db_repo,
		reshaper:       *NewReshaper(db_repo),
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

func (service *TransformService) RenderTemplate(ctx context.Context, template_name string, raw_data map[string]any) (string, error) {
	template_ := &db.Template{Name: template_name}
	err := service.db_repo.GetElementByName(ctx, template_)

	if err != nil {
		return "", err
	}

	requiredTags := map[string]any{}
	repeatTags := map[string]string{}

	formatted_template := service.reshaper.TransformTemplate(ctx, template_.Content, &requiredTags, &repeatTags)

	form_template, err := template.New(template_name).Parse(formatted_template)

	if err != nil {
		return "", err
	}
	

	data := raw_data // TODO: remove

	return execute(form_template, data)
}
