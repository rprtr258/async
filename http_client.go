package imhttp

import (
	"net/http"
	"net/url"
)

type Result[T any] struct {
	Value T
	Error error
}

func Get(url url.URL) Future[Result[*http.Response]] {
	return NewFuture(func() Result[*http.Response] {
		resp, err := http.Get(url.String())
		return Result[*http.Response]{resp, err}
	})
}
