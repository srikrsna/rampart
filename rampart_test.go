package rampart_test

import (
	"context"
	"testing"
	"time"

	"github.com/srikrsna/rampart"
	"github.com/srikrsna/rampart/stores/memory"
)

func TestRampart(t *testing.T) {
	const limit = 15
	r := rampart.NewRampart(memory.NewMemoryStore(), time.Second, limit)
	run := func(t *testing.T, key string, iter int) bool {
		var (
			canpass bool
			err     error
		)
		for i := 0; i < iter; i++ {
			canpass, err = r.CanPass(context.TODO(), "key")
			if err != nil {
				t.Fatal(err)
			}

			if !canpass {
				break
			}
		}

		return canpass
	}
	t.Run("Pass", func(t *testing.T) {
		canpass := run(t, "pass", limit)
		if !canpass {
			t.Error("limit reached early")
		}
	})

	t.Run("Stop", func(t *testing.T) {
		canpass := run(t, "stop", limit+1)

		if canpass {
			t.Error("should be blocked")
		}
	})
}
