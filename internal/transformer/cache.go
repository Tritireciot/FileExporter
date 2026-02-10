package transformer

import (
	"sync"
	"html/template"
)

type Cache struct {
	mutex sync.RWMutex
	items map[string]*template.Template
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]*template.Template),
	}
}

func (cache *Cache) Get(name string) (*template.Template, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	template_, ok := cache.items[name]
	return template_, ok
}

func (cache *Cache) Set(name string, template_ *template.Template) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.items[name] = template_
}

func (cache *Cache) Delete(name string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	delete(cache.items, name)
}