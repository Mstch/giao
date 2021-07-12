package test

import (
	"testing"
	"time"
)

func TestAppend(t *testing.T) {
	s := make([]int, 0)
	s = append(s, 1)
	ti := time.NewTimer(1*time.Second)
	ti.Stop()
	ti.Stop()
}
