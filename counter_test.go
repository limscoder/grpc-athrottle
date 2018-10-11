package grpc_athrottle_test

import (
	"testing"
	"time"

	throttle "github.com/limscoder/grpc-athrottle"

	"github.com/stretchr/testify/require"
)

type counterParams struct {
	binCount    int
	binDuration time.Duration
	minReq      int
	k           float32
	seed        int64
	callback    func(throttle.CounterEvent)
}

var defaultParams = counterParams{
	binCount:    30,
	binDuration: 10 * time.Millisecond,
	minReq:      0,
	k:           2.,
	seed:        1,
	callback:    func(event throttle.CounterEvent) {},
}

func newCounter(params counterParams) *throttle.Counter {
	return throttle.NewCounter(params.binCount, params.binDuration, params.minReq, params.k, params.seed, params.callback)
}

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
		if c.RejectNext() {
			rejectCount++
		}
	}
	return rejectCount
}

func TestCounter_RespectsMinReq(t *testing.T) {
	params := counterParams{
		binCount:    30,
		binDuration: 10 * time.Millisecond,
		minReq:      100,
		k:           2.,
		seed:        1,
		callback:    func(event throttle.CounterEvent) {},
	}

	c := newCounter(params)
	markCount(c, 25, 75)
	require.Equal(t, 0, rejectCount(c, 100))

	c = newCounter(params)
	require.Equal(t, 0, rejectCount(c, 100))
}

func TestCounter_Rejects(t *testing.T) {
	c := newCounter(defaultParams)
	markCount(c, 25, 50)
	require.Equal(t, 0, rejectCount(c, 100))

	c = newCounter(defaultParams)
	markCount(c, 25, 75)
	require.Equal(t, 38, rejectCount(c, 100))

	c = newCounter(defaultParams)
	markCount(c, 25, 100)
	require.Equal(t, 51, rejectCount(c, 100))

	c = newCounter(defaultParams)
	markCount(c, 25, 500)
	require.Equal(t, 91, rejectCount(c, 100))
}

func TestCounter_Resets(t *testing.T) {
	params := counterParams{
		binCount:    5,
		binDuration: 10 * time.Millisecond,
		minReq:      0,
		k:           2.,
		seed:        1,
		callback:    func(event throttle.CounterEvent) {},
	}

	c := newCounter(params)
	markCount(c, 25, 75)
	require.Equal(t, 38, rejectCount(c, 100))

	time.Sleep(6 * 10 * time.Millisecond)
	require.Equal(t, 0, rejectCount(c, 100))
}

func TestCounter_Events(t *testing.T) {
	requestSum := 0
	acceptSum := 0
	rejectSum := 0
	callback := func(event throttle.CounterEvent) {
		if event == throttle.RequestEvent {
			requestSum++
		} else if event == throttle.AcceptEvent {
			acceptSum++
		} else if event == throttle.RejectEvent {
			rejectSum++
		}
	}

	c := newCounter(counterParams{
		binCount:    5,
		binDuration: 10 * time.Millisecond,
		minReq:      0,
		k:           2,
		seed:        1,
		callback:    callback,
	})
	markCount(c, 25, 75)
	rejectCount(c, 100)
	require.Equal(t, 75, requestSum)
	require.Equal(t, 25, acceptSum)
	require.Equal(t, 38, rejectSum)
}
