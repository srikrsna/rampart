package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/srikrsna/rampart"
)

var _ rampart.CounterStore = (*RedisStore)(nil)

type RedisStore struct {
	cli redis.Scripter
}

func NewRedisStore(ctx context.Context, cli redis.Scripter) (*RedisStore, error) {
	if err := incrScript.Load(ctx, cli).Err(); err != nil {
		return nil, fmt.Errorf("rampart: redis: unable to load counter store script: %w", err)
	}

	return &RedisStore{
		cli: cli,
	}, nil
}

func (r *RedisStore) Incr(ctx context.Context, incrKey, getKey string, expr int64) (int64, int64, error) {
	res, err := incrScript.Run(ctx, r.cli, []string{incrKey, getKey}, expr).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("redis: unable to increment: %w", err)
	}

	results := res.([]interface{})

	prev, ok := results[1].(int64)
	if !ok {
		prev = 0
	}

	return results[0].(int64), prev, nil
}

func (r *RedisStore) Clear(ctx context.Context, keyPrefix string) error {
	return clearScript.Run(ctx, r.cli, []string{"0"}, keyPrefix+"*").Err()
}

var incrScript = redis.NewScript(`
	local current = redis.call('INCR', KEYS[1])

	if current == 1 then
		redis.call('EXPIRE', KEYS[1], ARGV[1])
	end

	local previous, err = redis.pcall('GET', KEYS[2])
	if err then
		previous = 0
	end	

	return {current, previous}
`)

var clearScript = redis.NewScript(`local keys = redis.call('keys', ARGV[1]); for i=1,#keys,5000 do redis.call('del', unpack(keys, i, math.min(i+4999, #keys))) end; return keys;`)
