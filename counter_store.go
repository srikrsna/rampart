package rampart

import "context"

type CounterStore interface {
	// Incr increments the incrKey by 1 returning the incremented value and returns the
	// value at getKey without any modification.
	// If any of the keys are not found their value should be treated as 0.
	Incr(ctx context.Context, incrKey, getKey string, expr int64) (incr int64, prev int64, err error)
	// Clear clears all keys with the given prefix
	Clear(ctx context.Context, keyPrefix string) error
}
