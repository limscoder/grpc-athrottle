# grpc-athrottle
Adaptive throttling for golang gRPC clients

gRPC interceptor that throttles outgoing connections based on trailing success rate and safely manages retries as described in the [Google SRE Book](https://landing.google.com/sre/book/chapters/handling-overload.html).

## usage

```go
grpc.Dial(address, grpc.WithStreamInterceptor(
  grpc_athrottle.NewStreamClientInterceptor(&grpc_athrottle.DefaultOptions)))
```