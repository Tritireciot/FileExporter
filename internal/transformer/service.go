package transformer

import (
	"bytes"
	"context"
	"former/internal/db"
	"html/template"
)

type TransformService struct {
	db_repo *db.DBRepository
	template_cache *Cache
	reshaper Reshaper
}

func NewTransformService(db_repo *db.DBRepository, ) *TransformService {
	return &TransformService{
		db_repo: db_repo,
		template_cache: NewCache(),
		reshaper: *NewReshaper(),
	}
}

func (service *TransformService) SaveTemplate(ctx context.Context, template *db.Template) error {
	if err := service.db_repo.AddTemplate(ctx, template); err != nil {
		return err
	}

	service.template_cache.Delete(template.Name)
	return nil
}

func execute(template_ *template.Template, data map[string]any) (string, error) {
	var buffer bytes.Buffer
	err := template_.Execute(&buffer, data)
	return buffer.String(), err

}

func (service *TransformService) RenderTemplate(ctx context.Context, template_name string, data map[string]any) (string, error) {
	if form_template, ok := service.template_cache.Get(template_name); ok {

		return execute(form_template, data)
	}


	template_, err := service.db_repo.GetTemplateByName(ctx, template_name)

	if err != nil {
		return "", err
	}


	formatted_template := service.reshaper.TransformTemplate(template_.Content)

	form_template, err := template.New(template_name).Parse(formatted_template)

	if err != nil {
		return "", err
	}
	service.template_cache.Set(template_name, form_template)

	return execute(form_template, data)
}