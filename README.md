# grpc-athrottle

Adaptive throttling for golang gRPC clients

gRPC interceptor that throttles outgoing connections based on trailing accept rate as described in [Google SRE Book](https://landing.google.com/sre/book/chapters/handling-overload.html).

## usage

```go
import (
  throttle "github.com/limscoder/grpc-athrottle"
  "google.golang.org/grpc"
)

func connect() {
  opts := throttle.ThrottleOptions{
    // sliding window for the past n minutes, used to calculate throttle ratio
    WindowDuration:  3 * time.Minute,
    // number of requests that must be present within WindowDuration before throttling
    MinRequestCount: 25,
    // ratio of requests/successes that must be met before throttling begins
    // refered to as "K" in the Google SRE book
    MaxRatio:        2.,
    // callback function for counter events,
    // useful for logging throttle events and/or metrics
    Callback:        func(CounterEvent) {},
  }

  conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(
    throttle.NewClientUnaryInterceptor(&opts)))
}
```

example usage with event logger: https://github.com/limscoder/kubetastic/blob/master/cmd/randoui/main.go
