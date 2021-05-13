package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/srikrsna/rampart"
)

var _ rampart.CounterStore = (*MemoryStore)(nil)

type MemoryStore struct {
	sync.Mutex

	data map[string]int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: map[string]int64{},
	}
}

func (r *MemoryStore) Incr(ctx context.Context, incrKey, getKey string, expr int64) (int64, int64, error) {
	r.Lock()
	defer r.Unlock()

	r.data[incrKey] += 1
	return r.data[incrKey], r.data[getKey], nil
}

func (r *MemoryStore) Clear(ctx context.Context, keyPrefix string) error {
	r.Lock()
	defer r.Unlock()

	for k := range r.data {
		if strings.HasPrefix(k, keyPrefix) {
			delete(r.data, k)
		}
	}

	return nil
}
