package std

import "unsafe"

type Iovec struct {
	Base unsafe.Pointer
	Len  uint64
}

func IovecFromBytes(buf []byte) Iovec {
	return Iovec{Base: unsafe.Pointer(unsafe.SliceData(buf)), Len: uint64(len(buf))}
}

func IovecFromString(s string) Iovec {
	return Iovec{Base: unsafe.Pointer(unsafe.StringData(s)), Len: uint64(len(s))}
}
