# imhttp

Immediate Mode HTTP Server Library.

> Actually now it just sends all requests to single channel still spawning goroutine for each incoming request. Fair implementation would require to parse http from each incoming connection while handling chunking and exceptional cases which is quite hard for me to implement without premade libraries.

See [main.go](./cmd/main.go) for example usage.
