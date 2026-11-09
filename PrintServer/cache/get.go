package cache

func (cache *LRUCache) Get(key ContentKey) (*CompiledTemplate, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	element, ok := cache.items[key]
	if !ok {
		return nil, false
	}
	cache.orderedContent.MoveToFront(element)
	value := element.Value.(*cacheEntry).value
	return &value, true
}

func (cache *LRUCache) GetValidated(id TemplateKey) (*CompiledTemplate, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	key, ok := cache.latest[id]
	if !ok {
		return nil, false
	}
	element, ok := cache.items[key]
	if !ok {
		return nil, false
	}
	cache.orderedContent.MoveToFront(element)
	value := element.Value.(*cacheEntry).value
	return &value, true
}
