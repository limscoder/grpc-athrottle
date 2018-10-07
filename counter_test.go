package grpc_athrottle_test

import (
	"testing"
	"time"

	throttle "github.com/limscoder/grpc-athrottle"

	"github.com/stretchr/testify/require"
)

func markCount(c *throttle.Counter, acceptCount int, requestCount int) {
	for i := 0; i < acceptCount; i++ {
		c.MarkAccept()
	}
	for i := 0; i < requestCount; i++ {
		c.MarkRequest()
	}
}

func rejectCount(c *throttle.Counter, count int) int {
	rejectCount := 0
	for i := 0; i < count; i++ {
		if c.RejectNext(2) {
			rejectCount++
		}
	}
	return rejectCount
}

func TestCounter_RespectsMinReq(t *testing.T) {
	c := throttle.NewCounter(30, 10*time.Second, 100, 1)
	markCount(c, 25, 75)
	require.Equal(t, 0, rejectCount(c, 100))

	c = throttle.NewCounter(30, 10*time.Second, 100, 1)
	require.Equal(t, 0, rejectCount(c, 100))
}

func TestCounter_Rejects(t *testing.T) {
	c := throttle.NewCounter(30, 10*time.Second, 0, 1)
	markCount(c, 25, 50)
	require.Equal(t, 0, rejectCount(c, 100))

	c = throttle.NewCounter(30, 10*time.Second, 0, 1)
	markCount(c, 25, 75)
	require.Equal(t, 38, rejectCount(c, 100))

	c = throttle.NewCounter(30, 10*time.Second, 0, 1)
	markCount(c, 25, 100)
	require.Equal(t, 51, rejectCount(c, 100))

	c = throttle.NewCounter(30, 10*time.Second, 0, 1)
	markCount(c, 25, 500)
	require.Equal(t, 91, rejectCount(c, 100))
}

func TestCounter_Resets(t *testing.T) {
	c := throttle.NewCounter(5, 10*time.Millisecond, 0, 1)
	markCount(c, 25, 75)
	require.Equal(t, 38, rejectCount(c, 100))

	time.Sleep(6 * 10 * time.Millisecond)
	require.Equal(t, 0, rejectCount(c, 100))
}
