package cache

import (
	"container/list"
	"errors"
	"html/template"
	"sync"
)

type ContentKey struct {
    Subsystem   string
    TemplateType string
    Hash        string
}

type TemplateKey struct {
    Subsystem string
    ID        int64
}

type LRUCache struct {
	mutex sync.Mutex
	limits CacheLimits
	currentBytes int64

	orderedContent *list.List
	items map[ContentKey]*list.Element
	latest map[TemplateKey]ContentKey
}

func NewLRUCache(limits CacheLimits) (*LRUCache, error) {
	if limits.MaxEntries <= 0 || limits.MaxBytes <= 0 {
        return nil, errors.New("лимиты кэша должны быть положительными")
    }

	return &LRUCache{
		limits: limits,
		orderedContent: list.New(),
		items: make(map[ContentKey]*list.Element),
		latest: make(map[TemplateKey]ContentKey),
	}, nil
}

type CompiledTemplate struct {
	Template *template.Template
	Size int64
}

type TemplateCache interface{
	Put(key ContentKey, compiledTemplate *CompiledTemplate) error
	PutValidated(id TemplateKey, key ContentKey, value *CompiledTemplate) bool
	Get(key ContentKey) (string, bool)
}