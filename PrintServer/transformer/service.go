package transformer

import (
	"PrintServer/PrintServer/cache"
	template_compiler "PrintServer/PrintServer/compiler"
	"PrintServer/PrintServer/config"
	"PrintServer/PrintServer/diagnostic"
	logging "PrintServer/agent"
	"context"
	"html/template"
	"net/http"
)

type Transformer interface {
	RenderTemplate(ctx context.Context, templateID int64, subsystem string, rawData any) (string, string, error)
	PreviewTemplate(ctx context.Context, templateContent string, subsystem string, templateType string) (string, string, error)
	ValidateTemplate(templateID int64, templateContent string, subsystem string, templateType string, contentHash string, templateData map[string]any) error
	PDFFromTemplate(htmlContent string) ([]byte, error)
	DOCXFromTemplate(htmlContent string) ([]byte, error)
}

type TransformService struct {
	parser  template_compiler.Parser
	pdfExec string
	client  *configClient

	mainCache    *cache.LRUCache
	previewCache *cache.LRUCache
}

func NewTransformService() *TransformService {
	configAddress := config.GetSubAccess().GetPath()
	printContentAddress := configAddress
	if config.PrintContentHost != "" && config.PrintContentPort != "" {
		printContentAddress = config.PrintContentHost + ":" + config.PrintContentPort
	}

	cacheConfig := cache.DefaultCacheConfig()

	if config.PrintCacheMainMaxEntries > 0 {
		cacheConfig.Main.MaxEntries = config.PrintCacheMainMaxEntries
	}

	if config.PrintCacheMainMaxBytes > 0 {
		cacheConfig.Main.MaxBytes = config.PrintCacheMainMaxBytes
	}

	mainCache, _ := cache.NewLRUCache(cacheConfig.Main)
	previewCache, _ := cache.NewLRUCache(cacheConfig.Preview)

	return &TransformService{
		parser:       template_compiler.NewParser(),
		pdfExec:      "",
		client:       newConfigClient(configAddress, printContentAddress),
		mainCache:    mainCache,
		previewCache: previewCache,
	}
}

func (service *TransformService) RenderTemplate(ctx context.Context, templateID int64, subsystem string, rawData any) (string, string, error) {
	tmpl := Template{ID: templateID, Subsystem: subsystem}
	err := service.client.getTemplate(ctx, &tmpl)

	if err != nil {
		logging.Agent.AddSimpleError("Подготовка шаблона на печать", "Не удалось получить шаблон: "+err.Error())
		return "", "", err
	}
	contentKey := cache.ContentKey{Subsystem: subsystem, TemplateType: tmpl.Type, Hash: tmpl.ContentHash}
	// Check the upstream response before a cache hit can hide a mismatched revision.
	if err := verifyContentHash(tmpl.Content, tmpl.ContentHash); err != nil {
		return "", "", err
	}
	cached := service.cachedMainTemplate(cache.TemplateKey{Subsystem: subsystem, ID: templateID}, contentKey)
	var compiled *template.Template
	if cached != nil {
		compiled = cached.Template
	} else {
		compiled, err = service.compileTemplate(tmpl.Content, tmpl.Name)
		if err != nil {
			return "", "", err
		}
	}

	data, err := service.client.getPrintData(ctx, rawData, &tmpl)
	if err != nil {
		logging.Agent.AddSimpleInfo("Подготовка шаблона печати", err.Error())
		return "", "", err
	}
	result, err := execute(compiled, data)
	if err != nil {
		return tmpl.Name, result, err
	}
	if cached == nil {
		if err := service.saveInCache(tmpl, compiled); err != nil {
			logging.Agent.AddSimpleError("Подготовка шаблона на печать", "Не удалось сохранить шаблон в кэше: "+err.Error())
			return "", "", err
		}
	}
	return tmpl.Name, result, nil
}

func (service *TransformService) PreviewTemplate(ctx context.Context, templateContent string, subsystem string, templateType string) (string, string, error) {
	var compiled *template.Template
	cached := true
	contentKey := cache.ContentKey{
		Subsystem:    subsystem,
		TemplateType: templateType,
		Hash:         hashTemplateContent(templateContent),
	}
	if val, ok := service.previewCache.Get(contentKey); ok {
		compiled = val.Template
	} else {
		var err error
		compiled, err = service.compileTemplate(templateContent, templateType)
		if err != nil {
			return "", "", err
		}
		cached = false
	}
	data, err := service.client.getPreviewData(ctx, subsystem, templateType)
	if err != nil {
		logging.Agent.AddSimpleInfo("Предпросмотр шаблон", err.Error())
		return "", "", err
	}
	result, err := execute(compiled, data)
	if err != nil {
		return "", "", err
	}

	if !cached {
		err = service.savePreviewInCache(contentKey, compiled, templateContent)
		if err != nil {
			logging.Agent.AddSimpleError("Предпросмотр шаблона", "Не удалось сохранить шаблон в кэше: "+err.Error())
			return "", "", err
		}
	}
	return result, templateType, nil
}

func (service *TransformService) ValidateTemplate(
	templateID int64,
	templateContent string,
	subsystem string,
	templateType string,
	contentHash string,
	templateData map[string]any,
) error {
	var compiled *template.Template
	templateKey := cache.TemplateKey{Subsystem: subsystem, ID: templateID}
	contentKey := cache.ContentKey{Subsystem: subsystem, TemplateType: templateType, Hash: contentHash}
	if err := verifyContentHash(templateContent, contentHash); err != nil {
		return diagnostic.NewRequestError(http.StatusUnprocessableEntity, diagnostic.CodeContentHashMismatch, "Хеш не соответствует содержимому шаблона")
	}

	cached := service.cachedMainTemplate(templateKey, contentKey)
	if cached != nil {
		compiled = cached.Template
	} else {
		var err error
		compiled, err = service.compileTemplate(templateContent, templateType)
		if err != nil {
			return err
		}
	}

	_, err := execute(compiled, templateData)

	if err != nil {
		return err
	}

	if cached == nil {
		if err := service.saveInCache(
			Template{
				ID:          templateID,
				Subsystem:   subsystem,
				Type:        templateType,
				ContentHash: contentHash,
				Content:     templateContent,
			},
			compiled,
		); err != nil {
			logging.Agent.AddSimpleError("Валидация шаблона", "Не удалось сохранить шаблон в кэше: "+err.Error())
			return err
		}
	}

	return nil
}
