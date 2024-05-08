package circular

import (
	"unsafe"

	. "sems/std"
)

const PageSize = 4096

type Buffer struct {
	Buf  []byte
	Head int
	Tail int
}

const (
	/* See <sys/mman.h>. */
	PROT_NONE  = 0x00
	PROT_READ  = 0x01
	PROT_WRITE = 0x02

	MAP_SHARED  = 0x0001
	MAP_PRIVATE = 0x0002

	MAP_FIXED = 0x0010
	MAP_ANON  = 0x1000
)

var (
	SHM_ANON = unsafe.String((*byte)(unsafe.Pointer(uintptr(1))), 8)
	NULL     = unsafe.String(nil, 0)
)

func New(pages int) (Buffer, error) {
	fd, err := ShmOpen2(SHM_ANON, O_RDWR, 0, 0, NULL)
	if err != nil {
		return Buffer{}, err
	}

	size := pages * PageSize

	if err := Ftruncate(fd, int64(size)); err != nil {
		return Buffer{}, err
	}

	buffer, err := Mmap(nil, 2*uint64(size), PROT_NONE, MAP_PRIVATE|MAP_ANON, -1, 0)
	if err != nil {
		return Buffer{}, err
	}

	if _, err := Mmap(buffer, uint64(size), PROT_READ|PROT_WRITE, MAP_SHARED|MAP_FIXED, fd, 0); err != nil {
		return Buffer{}, err
	}
	if _, err := Mmap(unsafe.Add(buffer, size), uint64(size), PROT_READ|PROT_WRITE, MAP_SHARED|MAP_FIXED, fd, 0); err != nil {
		return Buffer{}, err
	}

	cb := Buffer{
		Buf: unsafe.Slice((*byte)(buffer), 2*size),
	}
	// NOTE: sanity checks
	cb.Buf[0] = '\x00'
	cb.Buf[size-1] = '\x00'
	cb.Buf[size] = '\x00'
	cb.Buf[2*size-1] = '\x00'

	return cb, nil
}

func (cb *Buffer) Consume(n int) {
	cb.Head += n
	if cb.Head > len(cb.Buf)/2 {
		cb.Head -= len(cb.Buf) / 2
		cb.Tail -= len(cb.Buf) / 2
	}
}

func (cb *Buffer) Produce(n int) {
	cb.Tail += n
}

func (cb *Buffer) RemainingSlice() []byte {
	return cb.Buf[cb.Tail : cb.Head+len(cb.Buf)/2]
}

func (cb *Buffer) RemainingSpace() int {
	return (len(cb.Buf) / 2) - (cb.Tail - cb.Head)
}

func (cb *Buffer) Reset() {
	cb.Head = 0
	cb.Tail = 0
}

func (cb *Buffer) UnconsumedLen() int {
	return cb.Tail - cb.Head
}

func (cb *Buffer) UnconsumedSlice() []byte {
	return unsafe.Slice(&cb.Buf[cb.Head], cb.UnconsumedLen())
}

func (cb *Buffer) UnconsumedString() string {
	return unsafe.String(&cb.Buf[cb.Head], cb.UnconsumedLen())
}
