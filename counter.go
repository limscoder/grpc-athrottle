package grpc_athrottle

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type bin struct {
	acceptCount  int
	requestCount int
}

type Counter struct {
	sync.Mutex
	bins   []bin
	binIdx int
	minReq int
	rdm    *rand.Rand
}

func NewCounter(binCount int, binDuration time.Duration, minReq int, seed int) *Counter {
	c := &Counter{
		bins:   make([]bin, binCount, binCount),
		binIdx: 0,
		minReq: minReq,
		rdm:    rand.New(rand.NewSource(int64(seed))),
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

func (c *Counter) MarkRequest() {
	c.Lock()
	defer c.Unlock()
	c.bins[c.binIdx].requestCount++
}

func (c *Counter) MarkAccept() {
	c.Lock()
	defer c.Unlock()
	c.bins[c.binIdx].acceptCount++
}

func (c *Counter) RejectNext(k int) bool {
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

	p := math.Max(0., float64(requestCount-k*acceptCount)/float64(requestCount+1))
	return c.rdm.Float64() < p
}
