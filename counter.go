package grpc_athrottle

import (
	"math/rand"
	"sync"
	"time"
)

type bin struct {
	acceptCount  int
	requestCount int
}

// CounterEvent event code
type CounterEvent uint32

const (
	// RequestEvent is fired when a new request is dispatched
	RequestEvent CounterEvent = 0
	// AcceptEvent is fired when a request is accepted
	AcceptEvent CounterEvent = 1
	// RejectEvent is fired when a request is rejected
	RejectEvent CounterEvent = 2
)

// Counter tracks request status over a sliding window
type Counter struct {
	sync.Mutex
	bins     []bin
	binIdx   int
	minReq   int
	k        float32
	rdm      *rand.Rand
	callback func(CounterEvent)
}

// NewCounter returns Counter instance
func NewCounter(binCount int, binDuration time.Duration, minReq int, k float32, seed int64, callback func(CounterEvent)) *Counter {
	c := &Counter{
		bins:     make([]bin, binCount, binCount),
		binIdx:   0,
		minReq:   minReq,
		k:        k,
		rdm:      rand.New(rand.NewSource(seed)),
		callback: callback,
	}

	ticker := time.NewTicker(binDuration)
	go func() {
		lastIdx := binCount - 1
		for range ticker.C {
			c.Lock()
			nextIdx := 0
			if c.binIdx < lastIdx {
				nextIdx = c.binIdx + 1
			}

			c.binIdx = nextIdx
			c.bins[nextIdx].acceptCount = 0
			c.bins[nextIdx].requestCount = 0

			c.Unlock()
		}
	}()
	return c
}

// MarkRequest increments the total request count
func (c *Counter) MarkRequest() {
	c.Lock()
	defer c.Unlock()
	c.bins[c.binIdx].requestCount++
	c.callback(RequestEvent)
}

// MarkAccept increments the accepted request count
func (c *Counter) MarkAccept() {
	c.Lock()
	defer c.Unlock()
	c.bins[c.binIdx].acceptCount++
	c.callback(AcceptEvent)
}

// RejectNext returns true if a request should be rejected
func (c *Counter) RejectNext() bool {
	acceptCount := 0
	requestCount := 0

	c.Lock()
	defer c.Unlock()
	for _, bin := range c.bins {
		acceptCount += bin.acceptCount
		requestCount += bin.requestCount
	}

	if requestCount < c.minReq {
		return false
	}

	inflightCount := float32(requestCount) - c.k*float32(acceptCount)
	failRatio := inflightCount / float32(requestCount+1)
	reject := c.rdm.Float32() < failRatio
	if reject {
		c.callback(RejectEvent)
	}
	return reject
}
