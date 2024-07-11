package imhttp

import (
	"net/http"
	"net/url"
)

func Get(url url.URL) Promise[Result[*http.Response]] {
	return NewAsync(func() Result[*http.Response] {
		resp, err := http.Get(url.String())
		return Result[*http.Response]{resp, err}
	})
}
