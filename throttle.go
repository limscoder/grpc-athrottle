package grpc_athrottle

import (
	"context"
	"time"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc"
)

// ThrottleOptions contains options for the adaptive throttling logic.
type ThrottleOptions struct {
	WindowDuration  time.Duration      // sliding window used to calculate throttle ratio
	MinRequestCount int                // number of requests that must be present within WindowDuration before throttling
	MaxRatio        float32            // ratio of requests/successes that must be met before throttling begins
	Callback        func(CounterEvent) // Callback function for counter events, useful for logging throttle events and metrics
}

// DefaultOptions provides recommended defaults for ThrottleOptions
var DefaultOptions = ThrottleOptions{
	WindowDuration:  3 * time.Minute,
	MinRequestCount: 25,
	MaxRatio:        2.,
	Callback:        func(CounterEvent) {},
}

// DO NOT increment accept count when error response indicates server-side rejection or overloading
// DO increment accept count for normal application errors
var overloadCodes = []codes.Code{
	codes.Aborted,
	codes.FailedPrecondition,
	codes.Unavailable,
	codes.DeadlineExceeded,
	codes.ResourceExhausted,
}

func shouldAccept(err error) bool {
	if err == nil {
		return true
	}

	s, ok := status.FromError(err)
	if !ok {
		// unknown status code
		return false
	}

	c := s.Code()
	for i := 0; i < len(overloadCodes); i++ {
		if c == overloadCodes[i] {
			return false
		}
	}

	return true
}

func newCounter(throttleOpts *ThrottleOptions) *Counter {
	if throttleOpts == nil {
		throttleOpts = &DefaultOptions
	}

	binDuration := 10 * time.Second
	binCount := int(throttleOpts.WindowDuration / binDuration)
	return NewCounter(
		binCount,
		binDuration,
		throttleOpts.MinRequestCount,
		throttleOpts.MaxRatio,
		time.Now().UnixNano(),
		throttleOpts.Callback)
}

// NewUnaryClientInterceptor returns Interceptor that throttles client requests
// based on adaptive throttling model as described in the Google SRE book: https://landing.google.com/sre/book/chapters/handling-overload.html
func NewClientUnaryInterceptor(throttleOpts *ThrottleOptions, opts ...grpc.CallOption) grpc.UnaryClientInterceptor {
	counter := newCounter(throttleOpts)
	return func(parentCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if counter.RejectNext() {
			return grpc.Errorf(codes.ResourceExhausted, "client request throttled")
		}

		counter.MarkRequest()
		err := invoker(parentCtx, method, req, reply, cc, opts...)
		if shouldAccept(err) {
			counter.MarkAccept()
		}
		return err
	}
}

// NewStreamClientInterceptor returns Interceptor that throttles client requests
// based on adaptive throttling model as described in the Google SRE book: https://landing.google.com/sre/book/chapters/handling-overload.html
func NewClientStreamInterceptor(throttleOpts *ThrottleOptions, opts ...grpc.CallOption) grpc.StreamClientInterceptor {
	counter := newCounter(throttleOpts)
	return func(parentCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if counter.RejectNext() {
			return nil, grpc.Errorf(codes.ResourceExhausted, "client request throttled")
		}

		counter.MarkRequest()
		s, err := streamer(parentCtx, desc, cc, method, opts...)
		if shouldAccept(err) {
			counter.MarkAccept()
		}
		return s, err
	}
}
