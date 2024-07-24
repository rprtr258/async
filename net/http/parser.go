package http

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"unsafe"
)

type Arena struct {
	Buffer []byte
}

func (a *Arena) Alloc(n int) []byte {
	if len(a.Buffer)+n >= cap(a.Buffer) {
		buffer := make([]byte, max(len(a.Buffer)+n*2, cap(a.Buffer)*2))
		copy(buffer, a.Buffer)
		a.Buffer = buffer
	}
	return a.Buffer[len(a.Buffer)-n:]
}

func (a *Arena) Dealloc() {
	a.Buffer = a.Buffer[:0]
}

type arena[T any] struct {
	Buffer []T
}

func (a *arena[T]) Alloc() *T {
	res := new(T)
	a.Buffer = append(a.Buffer, *res)
	return res
}

func (a *arena[T]) Dealloc() {
	a.Buffer = a.Buffer[:0]
}

type HTTPRequest struct {
	Arena Arena

	Address string

	Method  string
	URL     url.URL
	Version string

	Headers map[string]string
	Body    []byte

	Form url.Values
}

func Parse(request string, r *HTTPRequest) error {

	type HTTPState int

	const (
		HTTPStateMethod HTTPState = iota
		HTTPStateURI
		HTTPStateVersion
		HTTPStateHeader
		HTTPStateBody

		HTTPStateUnknown
		HTTPStateDone
	)

	var ContentLength int

	pos := 0
	for State := HTTPStateMethod; State != HTTPStateDone; {
		switch State {
		default:
			log.Panicf("Unknown HTTP parser state %d", State)
		case HTTPStateUnknown:
			if len(request[pos:]) < 2 {
				return nil
			}
			if request[pos:pos+2] == "\r\n" {
				pos += len("\r\n")

				if ContentLength != 0 {
					State = HTTPStateBody
				} else {
					State = HTTPStateDone
				}
			} else {
				State = HTTPStateHeader
			}

		case HTTPStateMethod:
			if len(request[pos:]) < 4 {
				return nil
			}
			switch request[pos : pos+4] {
			case "GET ":
				r.Method = "GET"
			case "POST":
				r.Method = "POST"
			default:
				return fmt.Errorf("Method not allowed")
			}
			pos += len(r.Method) + 1
			State = HTTPStateURI
		case HTTPStateURI:
			lineEnd := strings.IndexByte(request[pos:], '\r')
			if lineEnd == -1 {
				return nil
			}

			uriEnd := strings.IndexByte(request[pos:pos+lineEnd], ' ')
			if uriEnd == -1 {
				return fmt.Errorf("Bad Request")
			}

			queryStart := strings.IndexByte(request[pos:pos+lineEnd], '?')
			if queryStart != -1 {
				r.URL.Path = request[pos : pos+queryStart]
				r.URL.RawQuery = request[pos+queryStart+1 : pos+uriEnd]
			} else {
				r.URL.Path = request[pos : pos+uriEnd]
				r.URL.RawQuery = ""
			}

			const httpVersionPrefix = "HTTP/"
			httpVersion := request[pos+uriEnd+1 : pos+lineEnd]
			if httpVersion[:len(httpVersionPrefix)] != httpVersionPrefix {
				return fmt.Errorf("Bad Request")
			}
			r.Version = httpVersion[len(httpVersionPrefix):]
			pos += len(r.URL.Path) + len(r.URL.RawQuery) + 1 + len(httpVersionPrefix) + len(r.Version) + len("\r\n")
			State = HTTPStateUnknown
		case HTTPStateHeader:
			lineEnd := strings.IndexByte(request[pos:], '\r')
			if lineEnd == -1 {
				return nil
			}
			k, v, _ := strings.Cut(request[pos:pos+lineEnd], " ")
			r.Headers[k] = v
			pos += lineEnd + len("\r\n")

			if k == "Content-Length:" {
				var err error
				ContentLength, err = strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("Bad Request")
				}
			}

			State = HTTPStateUnknown
		case HTTPStateBody:
			if len(request[pos:]) < ContentLength {
				return nil
			}

			r.Body = unsafe.Slice(unsafe.StringData(request[pos:]), ContentLength)
			pos += len(r.Body)
			State = HTTPStateDone
		}
	}

	return nil
}

func (r *HTTPRequest) Cookie(name string) string {
	for header, cookie := range r.Headers {
		if header == "Cookie:" {
			if strings.HasPrefix(cookie, name) {
				cookie = cookie[len(name):]
				if cookie[0] != '=' {
					return ""
				}
				return cookie[1:]
			}

		}
	}

	return ""
}

func (r *HTTPRequest) ParseForm() error {

	if len(r.Form) != 0 {
		return nil
	}

	var err error
	query := unsafe.String(unsafe.SliceData(r.Body), len(r.Body))
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")

		keyBuffer := r.Arena.Alloc(len(key))
		n, ok := URLDecode(keyBuffer, key)
		if !ok {
			if err == nil {
				err = fmt.Errorf("invalid key")
			}
			continue
		}
		key = unsafe.String(unsafe.SliceData(keyBuffer), n)

		valueBuffer := r.Arena.Alloc(len(value))
		n, ok = URLDecode(valueBuffer, value)
		if !ok {
			if err == nil {
				err = fmt.Errorf("invalid value")
			}
			continue
		}
		value = unsafe.String(unsafe.SliceData(valueBuffer), n)

		r.Form.Add(key, value)
	}

	return err
}

func URLDecode(decoded []byte, encoded string) (int, bool) {
	var n int
	for i := 0; i < len(encoded); i++ {
		if encoded[i] == '%' {
			hi, ok := ParseHexDigit(encoded[i+1])
			if !ok {
				return 0, false
			}

			lo, ok := ParseHexDigit(encoded[i+2])
			if !ok {
				return 0, false
			}

			decoded[n] = byte(hi<<4 | lo)
			i += 2
		} else if encoded[i] == '+' {
			decoded[n] = ' '
		} else {
			decoded[n] = encoded[i]
		}
		n++
	}
	return n, true
}

func ParseHexDigit(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'A' && c <= 'F':
		return 10 + c - 'A', true
	default:
		return 0, false
	}
}
