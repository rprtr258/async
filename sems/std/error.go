package std

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"syscall"
)

type Error struct {
	Message string
}

type ErrorWithCode struct {
	Message string
	Code    int
}

type PanicError struct {
	Value any
	Trace []byte
}

const (
	/* From <errno.h>. */
	ENOENT      = 2      /* No such file or directory */
	EINTR       = 4      /* Interrupted system call */
	EEXIST      = 17     /* File exists */
	EPIPE       = 32     /* Broken pipe */
	EAGAIN      = 35     /* Resource temporarily unavailable */
	EWOULDBLOCK = EAGAIN /* Operation would block */
	EINPROGRESS = 36     /* Operation now in progress */
	EOPNOTSUPP  = 45     /* Operation not supported */
	ECONNRESET  = 54     /* Connection reset by peer */
	ENOSYS      = 78     /* Function not implemented */
)

func NewError(msg string) Error {
	return Error{Message: msg}
}

func (e Error) Error() string {
	return e.Message
}

func NewErrorWithCode(msg string, code int) ErrorWithCode {
	return ErrorWithCode{Message: msg, Code: code}
}

func (e ErrorWithCode) Error() string {
	buffer := make([]byte, 512)
	n := copy(buffer, e.Message)
	buffer[n] = ' '
	n++

	if e.Code != 0 {
		n += SlicePutInt(buffer[n:], e.Code)
	}

	return string(buffer[:n])
}

func NewPanicError(value any) PanicError {
	return PanicError{Value: value, Trace: debug.Stack()}
}

func (e PanicError) Error() string {
	buffer := make([]byte, 0, 1024)
	buffer = fmt.Appendf(buffer, "%v\n", e.Value)
	buffer = append(buffer, e.Trace...)
	return string(buffer)
}

func NewSyscallError(msg string, errno syscall.Errno) error {
	if errno == 0 {
		return nil
	}
	return error(ErrorWithCode{Message: msg + ": " + errno.Error(), Code: int(errno)})
}

func WrapErrorWithTrace(err error, skip int) error {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return err
	}
	return fmt.Errorf("%s:%d: %w", file, line, err)
}
