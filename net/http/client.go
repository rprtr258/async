package http

import (
	"net/http"
	"net/url"

	. "github.com/rprtr258/async"
)

func Get(url url.URL) Future[Result[*http.Response]] {
	return NewFuture(func() Result[*http.Response] {
		resp, err := http.Get(url.String())
		return Result[*http.Response]{resp, err}
	})
}
