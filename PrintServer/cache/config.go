package cache

type CacheLimits struct {
	MaxEntries int
	MaxBytes int64
}

type CacheConfig struct {
	Main CacheLimits
	Preview CacheLimits
}

func DefaultCacheConfig() CacheConfig {
	const mib = 1 << 20

	return CacheConfig{
		Main: CacheLimits{MaxEntries: 1024, MaxBytes: 100 * mib},
		Preview: CacheLimits{MaxEntries: 32, MaxBytes: 8 * mib},
	}
}