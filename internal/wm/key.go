package wm

const (
	fnv1a64Offset = 14695981039346656037
	fnv1a64Prime  = 1099511628211
)

func SeedFromKey(key string) uint64 {
	hash := uint64(fnv1a64Offset)
	for i := 0; i < len(key); i++ {
		hash ^= uint64(key[i])
		hash *= fnv1a64Prime
	}
	return hash
}
