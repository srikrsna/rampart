package tests

import (
	"context"
	"testing"
	"time"

	"github.com/srikrsna/rampart"
)

func TestCounter(t *testing.T, str rampart.CounterStore) {
	const prefix = "prf"
	expiry := int64(time.Minute.Seconds())

	if err := str.Clear(context.TODO(), prefix); err != nil {
		t.Error(err)
	}

	cur, prev, err := str.Incr(context.TODO(), prefix+"1", prefix+"2", expiry)
	if err != nil {
		t.Fatal(err)
	}

	if cur != 1 {
		t.Errorf("incr should be 1 for a new key, got: %d", cur)
	}

	if prev != 0 {
		t.Errorf("prev should be 0 for a new key, got: %d", prev)
	}

	cur, prev, err = str.Incr(context.TODO(), prefix+"1", prefix+"2", expiry)
	if err != nil {
		t.Fatal(err)
	}

	if cur != 2 {
		t.Errorf("incr should be 2 for single increment, got: %d", cur)
	}

	if prev != 0 {
		t.Errorf("prev should be 0 for a new key, got: %d", prev)
	}
}
