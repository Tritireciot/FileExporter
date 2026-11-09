package cache_test

import (
	cache "PrintServer/PrintServer/cache"
	"fmt"
	"html/template"
	"math"
	"sync"
	"testing"
)

func newCache(t *testing.T, entries int, bytes int64) *cache.LRUCache {
	t.Helper()
	value, err := cache.NewLRUCache(cache.CacheLimits{MaxEntries: entries, MaxBytes: bytes})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func content(hash string) cache.ContentKey {
	return cache.ContentKey{Subsystem: "news", TemplateType: "rundown", Hash: hash}
}

func compiled(key cache.ContentKey, size int64) *cache.CompiledTemplate {
	return &cache.CompiledTemplate{
		Template: template.Must(template.New(key.Hash).Parse("ready")),
		Size:     size,
		Key:      key,
	}
}

func requirePresent(t *testing.T, c *cache.LRUCache, key cache.ContentKey) *cache.CompiledTemplate {
	t.Helper()
	value, ok := c.Get(key)
	if !ok || value == nil || value.Key != key {
		t.Fatalf("expected cached content %v, got value=%+v present=%v", key, value, ok)
	}
	return value
}

func requireAbsent(t *testing.T, c *cache.LRUCache, key cache.ContentKey) {
	t.Helper()
	if value, ok := c.Get(key); ok || value != nil {
		t.Fatalf("expected %v to be absent, got %+v", key, value)
	}
}

func TestCacheConfigurationAndValidation(t *testing.T) {
	for _, limits := range []cache.CacheLimits{{}, {MaxEntries: -1, MaxBytes: 1}, {MaxEntries: 1, MaxBytes: 0}} {
		if value, err := cache.NewLRUCache(limits); err == nil || value != nil {
			t.Errorf("limits %+v should fail, got value=%v err=%v", limits, value, err)
		}
	}
	defaults := cache.DefaultCacheConfig()
	if defaults.Main.MaxEntries <= 0 || defaults.Main.MaxBytes <= 0 || defaults.Preview.MaxEntries <= 0 || defaults.Preview.MaxBytes <= 0 {
		t.Errorf("invalid default limits: %+v", defaults)
	}
	if (cache.ContentKey{}).Validate() || (cache.TemplateKey{}).Validate() {
		t.Error("empty keys must be invalid")
	}
	if !content("a").Validate() || !(cache.TemplateKey{Subsystem: "news", ID: 1}).Validate() {
		t.Error("complete keys must be valid")
	}
	if !(cache.ContentKey{Subsystem: "news", Hash: "revision"}).Validate() {
		t.Error("an empty template type must not prevent caching a valid revision")
	}

	c := newCache(t, 2, 10)
	key := content("a")
	for _, tc := range []struct {
		name  string
		key   cache.ContentKey
		value *cache.CompiledTemplate
	}{
		{"empty key", cache.ContentKey{}, compiled(content("a"), 1)},
		{"nil value", key, nil},
		{"nil template", key, &cache.CompiledTemplate{Key: key, Size: 1}},
		{"zero size", key, compiled(key, 0)},
		{"negative size", key, compiled(key, -1)},
		{"mismatched key", key, compiled(content("b"), 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := c.Put(tc.key, tc.value); err == nil {
				t.Error("expected validation error")
			}
			requireAbsent(t, c, key)
		})
	}
	if err := c.PutValidated(cache.TemplateKey{}, key, compiled(key, 1)); err == nil {
		t.Error("invalid template ID was accepted")
	}
	if err := c.PutValidated(cache.TemplateKey{Subsystem: "other", ID: 1}, key, compiled(key, 1)); err == nil {
		t.Error("mismatched subsystem was accepted")
	}
}

func TestCacheLRUAndLimits(t *testing.T) {
	a, b, cKey := content("a"), content("b"), content("c")
	byEntries := newCache(t, 2, 100)
	for _, key := range []cache.ContentKey{a, b} {
		if err := byEntries.Put(key, compiled(key, 1)); err != nil {
			t.Fatal(err)
		}
	}
	requirePresent(t, byEntries, a) // A is now the newest entry.
	if err := byEntries.Put(cKey, compiled(cKey, 1)); err != nil {
		t.Fatal(err)
	}
	requireAbsent(t, byEntries, b)
	requirePresent(t, byEntries, a)
	requirePresent(t, byEntries, cKey)

	byBytes := newCache(t, 3, 5)
	if err := byBytes.Put(a, compiled(a, 3)); err != nil {
		t.Fatal(err)
	}
	if err := byBytes.Put(b, compiled(b, 3)); err != nil {
		t.Fatal(err)
	}
	requireAbsent(t, byBytes, a)
	requirePresent(t, byBytes, b)
	if err := byBytes.Put(cKey, compiled(cKey, 6)); err != nil {
		t.Fatalf("oversized content may be executed without caching: %v", err)
	}
	requireAbsent(t, byBytes, cKey)
	requirePresent(t, byBytes, b)
}

func TestCacheDoesNotOverflowByteLimit(t *testing.T) {
	c := newCache(t, 2, math.MaxInt64)
	a, b := content("a"), content("b")
	if err := c.Put(a, compiled(a, math.MaxInt64-1)); err != nil {
		t.Fatal(err)
	}
	if err := c.Put(b, compiled(b, 2)); err != nil {
		t.Fatal(err)
	}
	requireAbsent(t, c, a)
	requirePresent(t, c, b)
}

func TestCacheDeduplicatesAndCopiesWrapper(t *testing.T) {
	c := newCache(t, 1, 10)
	key := content("same")
	first := compiled(key, 3)
	if err := c.Put(key, first); err != nil {
		t.Fatal(err)
	}
	originalTemplate := first.Template
	first.Size = 99
	first.Key = content("changed")
	if err := c.Put(key, compiled(key, 5)); err != nil {
		t.Fatal(err)
	}
	got := requirePresent(t, c, key)
	if got.Size != 3 || got.Template != originalTemplate {
		t.Errorf("existing compiled entry changed: %+v", got)
	}
	got.Size = 0
	if requirePresent(t, c, key).Size != 3 {
		t.Error("mutating a returned wrapper changed the cached entry")
	}
}

func TestCacheValidatedVersionsAndSharedContent(t *testing.T) {
	c := newCache(t, 4, 100)
	id1 := cache.TemplateKey{Subsystem: "news", ID: 1}
	id2 := cache.TemplateKey{Subsystem: "news", ID: 2}
	a, b, next := content("a"), content("b"), content("next")
	for _, id := range []cache.TemplateKey{id1, id2} {
		if err := c.PutValidated(id, a, compiled(a, 1)); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.PutValidated(id1, b, compiled(b, 1)); err != nil {
		t.Fatal(err)
	}
	if got, ok := c.GetValidated(id1); !ok || got.Key != b {
		t.Errorf("id1 should point to B, got %+v %v", got, ok)
	}
	if got, ok := c.GetValidated(id2); !ok || got.Key != a {
		t.Errorf("id2 should still point to shared A, got %+v %v", got, ok)
	}
	if err := c.PutValidated(id2, next, compiled(next, 1)); err != nil {
		t.Fatal(err)
	}
	requireAbsent(t, c, a)
	if err := c.PutValidated(id1, content("huge"), compiled(content("huge"), 101)); err != nil {
		t.Fatal(err)
	}
	if got, ok := c.GetValidated(id1); ok || got != nil {
		t.Errorf("oversized new version left stale ID mapping: %+v %v", got, ok)
	}
	requireAbsent(t, c, b)
	requirePresent(t, c, next)
}

func TestCacheEvictionClearsValidatedMapping(t *testing.T) {
	c := newCache(t, 1, 10)
	id := cache.TemplateKey{Subsystem: "news", ID: 1}
	a, b := content("a"), content("b")
	if err := c.PutValidated(id, a, compiled(a, 1)); err != nil {
		t.Fatal(err)
	}
	if err := c.Put(b, compiled(b, 1)); err != nil {
		t.Fatal(err)
	}
	if value, ok := c.GetValidated(id); ok || value != nil {
		t.Errorf("evicted version remains accessible by ID: %+v %v", value, ok)
	}
}

func TestValidatedGetRefreshesLRUAndInvalidUpdatePreservesVersion(t *testing.T) {
	c := newCache(t, 2, 10)
	id := cache.TemplateKey{Subsystem: "news", ID: 1}
	a, b, next := content("a"), content("b"), content("next")
	if err := c.PutValidated(id, a, compiled(a, 1)); err != nil {
		t.Fatal(err)
	}
	if err := c.Put(b, compiled(b, 1)); err != nil {
		t.Fatal(err)
	}
	if value, ok := c.GetValidated(id); !ok || value.Key != a {
		t.Fatalf("expected validated A, got %+v %v", value, ok)
	}
	if err := c.Put(next, compiled(next, 1)); err != nil {
		t.Fatal(err)
	}
	requireAbsent(t, c, b)
	requirePresent(t, c, a)
	if err := c.PutValidated(id, next, nil); err == nil {
		t.Error("invalid update was accepted")
	}
	if value, ok := c.GetValidated(id); !ok || value.Key != a {
		t.Errorf("invalid update changed validated version: %+v %v", value, ok)
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	c := newCache(t, 8, 64)
	const workers, iterations = 24, 600
	keys := make([]cache.ContentKey, 16)
	values := make([]*cache.CompiledTemplate, len(keys))
	for i := range keys {
		keys[i] = content(fmt.Sprintf("hash-%d", i))
		values[i] = compiled(keys[i], int64(i%4+1))
	}
	var wg sync.WaitGroup
	errorsFound := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			id := cache.TemplateKey{Subsystem: "news", ID: int64(worker%6 + 1)}
			for step := 0; step < iterations; step++ {
				index := (worker*17 + step) % len(keys)
				key := keys[index]
				switch step % 4 {
				case 0:
					if err := c.Put(key, values[index]); err != nil {
						errorsFound <- err
						return
					}
				case 1:
					if err := c.PutValidated(id, key, values[index]); err != nil {
						errorsFound <- err
						return
					}
				case 2:
					if got, ok := c.Get(key); ok && (got == nil || got.Key != key || got.Size <= 0) {
						errorsFound <- fmt.Errorf("invalid Get result for %v: %+v", key, got)
						return
					}
				case 3:
					if got, ok := c.GetValidated(id); ok && (got == nil || got.Key.Subsystem != id.Subsystem || got.Size <= 0) {
						errorsFound <- fmt.Errorf("invalid GetValidated result for %v: %+v", id, got)
						return
					}
				}
			}
		}(worker)
	}
	wg.Wait()
	close(errorsFound)
	for err := range errorsFound {
		t.Error(err)
	}
}
