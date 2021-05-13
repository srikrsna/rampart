package memory_test

import (
	"testing"

	"github.com/srikrsna/rampart/stores/memory"
	"github.com/srikrsna/rampart/tests"
)

func TestRedisCounter(t *testing.T) {
	str := memory.NewMemoryStore()
	tests.TestCounter(t, str)
}
