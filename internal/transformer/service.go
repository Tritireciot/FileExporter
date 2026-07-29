package transformer

import (
	logging "PrintServer/agent"
	"PrintServer/internal/db"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)
type Transformer interface {
	RenderTemplate(ctx context.Context, template_id int, raw_data any) (string, string, error)
	PDFFromTemplate(htmlContent string) ([]byte, error)
	DOCXFromTemplate(htmlContent string) ([]byte, error)
}

type TransformService struct {
	db_repo  db.DBRepo
	reshaper Reshaper
	pdfExec  string
}

func setupPDFexecutor() (string, error) {
	tempDir := filepath.Join(os.TempDir(), "go-weasyprint-engine")
	var exePath string

	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err := unzipBytes(weasyprintZipBytes, tempDir); err != nil {
			return "", fmt.Errorf("ошибка распаковки движка: %v", err)
		}
	}

	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "weasyprint" {
			exePath = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("ошибка поиска бинарника: %v", err)
	}
	if exePath == "" {
		return exePath, fmt.Errorf("бинарник weasyprint не найден")
	}

	if err := os.Chmod(exePath, 0755); err != nil {
		return "", fmt.Errorf("ошибка прав chmod: %v", err)
	}

	return exePath, nil
}

func NewTransformService(db_repo db.DBRepo) *TransformService {
	pdfexec, err := setupPDFexecutor()
	if err != nil {
		logging.Agent.AddSimpleError("Создание TransformService", "Не удалось создать " + err.Error())
		return nil
	}
	return &TransformService{
		db_repo:  db_repo,
		reshaper: *NewReshaper(db_repo),
		pdfExec: pdfexec,
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

func framesToTime(frames int) string {
	fps := 25
	rest_frames := frames % fps
	seconds := frames / fps
	rest_seconds := seconds % 60
	minutes := seconds / 60
	rest_minutes := minutes % 60
	hours := minutes / 60
	return fmt.Sprintf("%02d:%02d:%02d:%02d", hours, rest_minutes, rest_seconds, rest_frames)
}

func filter(subsystem string, raw_data *any, render_data map[string]bool) {
	if subsystem == "news" && raw_data != nil {
		if rundown_data, ok := (*raw_data).(map[string]any); ok {
			
			rundown, ok := rundown_data["rundown"].(map[string]any)
			if !ok {
				return
			}

			for key := range rundown {
				logging.Agent.AddSimpleInfo("Info", key)
				if strings.Contains(key, "durat") {
					rundown[key] = framesToTime(int(rundown[key].(float64)))
				}
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
				story := object["story"].(map[string]any)
				for key := range story {
					if strings.Contains(key, "durat") {
						story[key] = framesToTime(int(story[key].(float64)))
					}
				}
				object["story"] = story
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
	logging.Agent.AddSimpleInfo("Подготовка шаблона печати", "Отфильтрованны данные")


	byte_data, err := json.Marshal(raw_data)
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
