package imhttp

import (
	"bytes"
	"fmt"
	"io"
)

const _bufferCapacity = 8 * 1024
const _userBufferCapacity = _bufferCapacity

func init() {
	assert(_bufferCapacity <= _userBufferCapacity, "The user buffer should be at least as big as the rolling buffer because sometimes you may wanna put the whole rollin content into the user buffer.")
}

type ImHTTP struct {
	socket io.ReadWriter

	buffer     [_bufferCapacity]byte
	bufferSize int

	userBuffer [_userBufferCapacity]byte

	contentLength int
	chunked       bool
	chunkedLength int
	chunkedDone   bool
}

func New(
	conn io.ReadWriter,
) *ImHTTP {
	return &ImHTTP{
		socket: conn,
	}
}

// TODO: not all methods are supported
type Method int

const (
	IMHTTP_GET Method = iota
	IMHTTP_POST
)

func (method Method) String() string {
	switch method {
	case IMHTTP_GET:
		return "GET"
	case IMHTTP_POST:
		return "POST"
	default:
		panic("unreachable")
	}
}

func (imhttp *ImHTTP) dropRollinBuffer(n int) {
	assert(n <= imhttp.bufferSize, "")
	copy(imhttp.buffer[:], imhttp.buffer[n:])
	imhttp.bufferSize -= n
}

func (imhttp *ImHTTP) shiftRollinBuffer(n int) []byte {
	// Boundary check
	/// assert(imhttp.rollin_buffer <= end)
	/// n := end - imhttp.rollin_buffer
	assert(n <= imhttp.bufferSize, "")

	// Copy chunk to user buffer
	assert(n <= _userBufferCapacity, "")
	copy(imhttp.userBuffer[:], imhttp.buffer[:n])

	// Shift buffer
	copy(imhttp.buffer[:], imhttp.buffer[n:imhttp.bufferSize])
	imhttp.bufferSize -= n

	return imhttp.userBuffer[:n]
}

// TODO: document that does not perform any reads until *everything* inside of rollin_buffer is processed
func (imhttp *ImHTTP) read() {
	if imhttp.bufferSize != 0 {
		return
	}

	n, _ := imhttp.socket.Read(imhttp.buffer[imhttp.bufferSize:])
	// TODO: handle read errors
	assert(n > 0, "")
	imhttp.bufferSize += n
}

func (imhttp *ImHTTP) bufferSlice() []byte {
	return imhttp.buffer[:imhttp.bufferSize]
}

func (imhttp *ImHTTP) writeString(s string) {
	// TODO: handle ImHTTP_Write errors
	_, _ = imhttp.socket.Write([]byte(s))
}

type req struct {
	*ImHTTP
}

func (imhttp *ImHTTP) Req(method Method, path string) *req {
	imhttp.writeString(method.String())
	imhttp.writeString(" ")
	// TODO: it is easy to make the resource malformed in imhttp_req_begin
	imhttp.writeString(path)
	imhttp.writeString(" HTTP/1.1\r\n")
	return &req{imhttp}
}

func (r *req) ReqHeaders(headers map[string]string) *req {
	for name, value := range headers {
		r.writeString(name)
		r.writeString(": ")
		r.writeString(value)
		r.writeString("\r\n")
	}
	r.writeString("\r\n")
	return r
}

func (r *req) ReqBodyChunk(chunk string) *req {
	r.writeString(chunk)
	return r
}

func (r *req) ReqBodyChunkBytes(chunk []byte) *req {
	_, _ = r.socket.Write(chunk)
	return r
}

func (imhttp *ImHTTP) ResBegin() {
	imhttp.contentLength = -1
	imhttp.chunked = false
	imhttp.chunkedLength = 0
	imhttp.chunkedDone = false
}

func (imhttp *ImHTTP) ResStatusCode() uint64 {
	imhttp.read()
	buffer := imhttp.bufferSlice()
	status_line, buffer, _ := bytes.Cut(buffer, []byte{'\n'})
	assert(bytes.HasSuffix(status_line, []byte("\r")), "status line did not fit in buffer")
	status_line = imhttp.shiftRollinBuffer(len(status_line) + 1)
	httpVersion, status, _ := bytes.Cut(status_line, []byte{' '})
	_ = httpVersion // TODO: HTTP version is skipped
	code, _, _ := bytes.Cut(status, []byte{' '})
	return parseU64(code)
}

// TODO: Document that this invalidate name and value on the consequent imhttp_* calls
func (imhttp *ImHTTP) ResNextHeader(name, value *[]byte) bool {
	imhttp.read()
	buffer := imhttp.bufferSlice()
	header_line, buffer, _ := bytes.Cut(buffer, []byte{'\n'})
	assert(bytes.HasSuffix(header_line, []byte("\r")), "header line did not fit in buffer")
	// Transfer the ownership of header_line from rollin_buffer to user_buffer
	header_line = imhttp.shiftRollinBuffer(len(header_line) + 1)

	if bytes.Equal(header_line, []byte("\r\n")) {
		return false
	}

	// TODO: don't set name/value if the user set them to NULL in imhttp_res_next_header
	*name, *value, _ = bytes.Cut(header_line, []byte(": "))
	*value, _, _ = bytes.Cut(*value, []byte{'\r', '\n'})

	// TODO: are header case-sensitive?
	if bytes.Equal(*name, []byte("Content-Length")) {
		// TODO: content_length overflow
		imhttp.contentLength = int(parseU64(*value))
	} else if bytes.Equal(*name, []byte("Transfer-Encoding")) {
		encoding_list := *value
		for len(encoding_list) > 0 {
			var encoding []byte
			encoding, encoding_list, _ = bytes.Cut(encoding_list, []byte(", "))
			if bytes.Equal(encoding, []byte("chunked")) {
				imhttp.chunked = true
			}
		}
	}
	return true
}

// TODO: document that the chunk is always invalidated after each call
func (imhttp *ImHTTP) ResNextBodyChunk(chunk *[]byte) bool {
	if imhttp.chunked {
		if !imhttp.chunkedDone {
			imhttp.read()

			if imhttp.chunkedLength == 0 {
				buffer := imhttp.bufferSlice()
				buffer, chunk_length_sv, _ := bytes.Cut(buffer, []byte{'\n'})
				assert(bytes.HasSuffix(chunk_length_sv, []byte("\r")), "chunk length did not fit in buffer")
				imhttp.chunkedLength = int(parseU64Hex(bytes.TrimSpace(chunk_length_sv)))
				imhttp.shiftRollinBuffer(len(chunk_length_sv) + 1)
			}

			if imhttp.chunkedLength == 0 {
				imhttp.chunkedDone = true
				return false
			}

			{
				n := imhttp.chunkedLength
				if n > imhttp.bufferSize {
					n = imhttp.bufferSize
				}

				data := imhttp.shiftRollinBuffer(n)
				imhttp.chunkedLength -= n

				if imhttp.chunkedLength == 0 {
					rollin := imhttp.bufferSlice()
					assert(bytes.HasPrefix(rollin, []byte("\r\n")), "")
					imhttp.dropRollinBuffer(2)
				}

				if chunk != nil {
					*chunk = data
				}
			}

			return true
		}
	} else {
		// TODO: ImHTTP can't handle the responses that do not set Content-Length
		assert(imhttp.contentLength >= 0, "can't handle the responses that do not set Content-Length")

		if imhttp.contentLength == 0 {
			return false
		}

		imhttp.read()
		buffer := imhttp.bufferSlice()
		// TODO: ImHTTP does not handle the situation when the server responded with more data than it claimed with Content-Length
		assert(
			len(buffer) <= imhttp.contentLength,
			fmt.Sprintf(
				"buffer size %d is larger than content length %d, buffer: %q",
				len(buffer), imhttp.contentLength, buffer))

		result := imhttp.shiftRollinBuffer(imhttp.bufferSize)

		if chunk != nil {
			*chunk = result
		}

		imhttp.contentLength -= len(result)

		return true
	}

	return false
}

func (*ImHTTP) ResEnd() {}
