package rampart

import (
	"context"
	"strconv"
	"time"
)

type Rampart struct {
	window int64
	expiry int64
	limit  int64

	str CounterStore
}

func NewRampart(str CounterStore, window time.Duration, limit int64) *Rampart {
	win := int64(window / time.Second)
	return &Rampart{
		window: win,
		expiry: 2*win + 2,
		limit:  limit,
		str:    str,
	}
}

func (r *Rampart) CanPass(ctx context.Context, key string) (bool, error) {
	now := time.Now().Unix()
	currentKeyTime := now - now%r.window
	previousKeyTime := currentKeyTime - r.window

	currentKey := key + strconv.FormatInt(currentKeyTime, 10)
	previousKey := key + strconv.FormatInt(previousKeyTime, 10)

	curr, prev, err := r.str.Incr(ctx, currentKey, previousKey, r.expiry)
	if err != nil {
		return false, err
	}

	rate := prev*(r.window-(now-currentKeyTime))/r.window + curr

	return rate <= r.limit, nil
}

func (r *Rampart) Clear(ctx context.Context, key string) error {
	return r.str.Clear(ctx, key)
}
