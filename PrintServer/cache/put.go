package cache

import (
	"container/list"
	"fmt"
)

type cacheEntry struct {
	key   ContentKey
	value CompiledTemplate
	size  int64
}

// Вызывается только под cache.mutex.
func (cache *LRUCache) deleteElementLocked(element *list.Element) {
	entry := element.Value.(*cacheEntry)

	cache.orderedContent.Remove(element)
	delete(cache.items, entry.key)
	cache.currentBytes -= entry.size

	for id, key := range cache.latest {
		if key == entry.key {
			delete(cache.latest, id)
		}
	}
}

// bool означает, что запись находится в кэше.
// Вызывается только под cache.mutex.
func (cache *LRUCache) putLocked(
	key ContentKey,
	value *CompiledTemplate,
) (bool, error) {
	if !key.Validate() {
		return false, fmt.Errorf("ключ шаблона пустой или неполный")
	}
	if value == nil || value.Template == nil || value.Size <= 0 {
		return false, fmt.Errorf("нет скомпилированного шаблона или некорректен его вес")
	}
	if key != value.Key {
		return false, fmt.Errorf("ключ записи не совпадает с ключом шаблона")
	}
	if element, ok := cache.items[key]; ok {
		cache.orderedContent.MoveToFront(element)
		return true, nil // Такой текст уже скомпилирован.
	}

	if value.Size > cache.limits.MaxBytes {
		return false, nil // Выполнить можно, хранить нельзя.
	}
	for len(cache.items) >= cache.limits.MaxEntries ||
		cache.currentBytes > cache.limits.MaxBytes-value.Size {
		cache.deleteElementLocked(cache.orderedContent.Back())
	}

	// The caller may reuse its wrapper after Put; keep a snapshot of its fields.
	entry := &cacheEntry{key: key, value: *value, size: value.Size}
	cache.items[key] = cache.orderedContent.PushFront(entry)
	cache.currentBytes += entry.size
	return true, nil
}

func (cache *LRUCache) Put(key ContentKey, value *CompiledTemplate) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	_, err := cache.putLocked(key, value)
	return err
}

func (cache *LRUCache) PutValidated(
	id TemplateKey,
	key ContentKey,
	value *CompiledTemplate,
) error {
	if !id.Validate() || id.Subsystem != key.Subsystem {
		return fmt.Errorf("некорректный идентификатор шаблона или подсистема ключа")
	}
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	oldKey, hadOld := cache.latest[id]

	stored, err := cache.putLocked(key, value)
	if err != nil {
		return err
	}

	if hadOld && oldKey != key {
		delete(cache.latest, id)

		// Не удаляем общий контент, если он нужен другому ID.
		usedByAnotherID := false
		for _, associatedKey := range cache.latest {
			if associatedKey == oldKey {
				usedByAnotherID = true
				break
			}
		}
		if !usedByAnotherID {
			if oldElement, ok := cache.items[oldKey]; ok {
				cache.deleteElementLocked(oldElement)
			}
		}
	}

	if stored {
		cache.latest[id] = key
	} else {
		// Новый шаблон слишком велик для кэша: старый ID не должен
		// продолжать указывать на предыдущую валидацию.
		delete(cache.latest, id)
	}
	return nil
}
