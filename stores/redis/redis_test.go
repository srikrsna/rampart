package redis_test

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	redisc "github.com/srikrsna/rampart/stores/redis"
	"github.com/srikrsna/rampart/tests"
)

func TestRedisCounter(t *testing.T) {

	cli := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	str, err := redisc.NewRedisStore(context.TODO(), cli)
	if err != nil {
		t.Fatal(err)
	}

	tests.TestCounter(t, str)
}
