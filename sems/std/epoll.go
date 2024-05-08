package std

import (
	"syscall"
)

func EpollCreate() (int, error) {
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return 0, NewSyscallError("kevent failed with code", err.(syscall.Errno))
	}
	return epfd, nil
}

const (
	OP_READ  = syscall.EPOLLIN
	OP_WRITE = syscall.EPOLLOUT
)

func EpollAdd(epfd int, fd FD, op uint32) error {
	err := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: op,
		Fd:     int32(fd),
		Pad:    0,
	})
	if err != nil {
		return NewSyscallError("kevent failed with code", err.(syscall.Errno))
	}
	return nil
}

func EpollWait(epfd int, events []syscall.EpollEvent) (int, error) {
	n, err := syscall.EpollWait(epfd, events, -1)
	if err != nil {
		return 0, NewSyscallError("wait failed with code", err.(syscall.Errno))
	}
	return n, nil
}
