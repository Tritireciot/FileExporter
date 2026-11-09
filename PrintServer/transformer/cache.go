package transformer

import (
	"PrintServer/PrintServer/cache"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
)

func countSize(content string) int64 {
	// реальный размер на деле другой этот средним является. Сделано в угоду скорости
	return 6 * 1024 + 8 * int64(len(content))
}

func hashTemplateContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func verifyContentHash(content, expected string) error {
	actual := hashTemplateContent(content)
	if actual != expected {
		return &TemplateIntegrityError{Expected: expected, Actual: actual}
	}
	return nil
}

func (service *TransformService) saveInCache(templateSummary Template, template *template.Template) error {
	contentKey := cache.ContentKey{
		Subsystem:    templateSummary.Subsystem,
		TemplateType: templateSummary.Type,
		Hash:         templateSummary.ContentHash,
	}
	ct := &cache.CompiledTemplate{
		Template: template,
		Key:      contentKey,
		Size: countSize(templateSummary.Content), 
	}
	err := service.mainCache.PutValidated(
		cache.TemplateKey{
			Subsystem: templateSummary.Subsystem,
			ID:        templateSummary.ID,
		},
		contentKey,
		ct,
	)
	return err
}

func (service *TransformService) cachedMainTemplate(id cache.TemplateKey, key cache.ContentKey) *cache.CompiledTemplate {
	if value, ok := service.mainCache.GetValidated(id); ok && value.Key == key {
		return value
	}
	value, _ := service.mainCache.Get(key)
	return value
}

func (service *TransformService) savePreviewInCache(key cache.ContentKey, compiledTemplate *template.Template, templateContent string) error {
	compiled := &cache.CompiledTemplate{Template: compiledTemplate, Key: key, Size: countSize(templateContent)}
	return service.previewCache.Put(key, compiled)
}
