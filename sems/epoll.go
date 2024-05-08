package main

import (
	"fmt"
	"syscall"
	"unsafe"

	. "sems/std"
)

type EventQueue struct {
	epfd int

	eventsBuffer [256]syscall.EpollEvent
	head         int
	tail         int
}

type EventRequest int

const (
	EventRequestRead EventRequest = 1 << iota
	EventRequestWrite
)

type EventType int32

const (
	EventNone EventType = iota
	EventRead
	EventWrite
	EventSignal
)

type Event struct {
	Type      EventType
	Fd        int
	Available int
	UserData  unsafe.Pointer

	EndOfFile bool
}

func NewEventQueue() (*EventQueue, error) {
	epfd, err := EpollCreate()
	if err != nil {
		return nil, fmt.Errorf("failed to open epoll: %w", err)
	}
	return &EventQueue{
		epfd: epfd,
	}, nil
}

func (q *EventQueue) AddSocket(fd FD, request EventRequest) error {
	if request&EventRequestRead != 0 {
		if err := EpollAdd(q.epfd, fd, OP_READ); err != nil {
			return fmt.Errorf("failed to request socket read event: %w", err)
		}
	}

	if request&EventRequestWrite != 0 {
		if err := EpollAdd(q.epfd, fd, OP_WRITE); err != nil {
			return fmt.Errorf("failed to request socket write event: %w", err)
		}
	}

	return nil
}

func (q *EventQueue) AddSignal(signals ...syscall.Signal) error {
	type sigset_t [2]uint32
	var sigs sigset_t
	for _, sig := range signals {
		sigs[(sig-1)/32] |= 1 << ((uint32(sig) - 1) & 31)
	}
	signal_fd, _, err := syscall.Syscall(syscall.SYS_SIGNALFD, ^uintptr(0), uintptr(unsafe.Pointer(&sigs)), syscall.O_NONBLOCK)
	if err != 0 {
		return fmt.Errorf("create signal fd: %w", err)
	}

	if err := EpollAdd(q.epfd, int(signal_fd), OP_READ); err != nil {
		return fmt.Errorf("failed to request signal event: %w", err)
	}

	return nil
}

func (q *EventQueue) Close() error {
	return Close(q.epfd)
}

func (q *EventQueue) GetEvent() (Event, error) {
	if q.head >= q.tail {
	retry:
		_, err := EpollWait(q.epfd, q.eventsBuffer[:])
		if err != nil {
			if err.(ErrorWithCode).Code == EINTR {
				goto retry
			}
			return Event{}, err
		}
		q.head = 0
	}
	head := q.eventsBuffer[q.head]
	q.head++

	var eventType EventType
	switch head.Events {
	case OP_READ:
		eventType = EventRead
	case OP_WRITE:
		eventType = EventWrite
		// case OP_SIGNAL:
		// 	eventType = EventSignal
	}

	return Event{
		Fd: int(head.Fd),
		// Available: head.Data,
		// UserData:  head.Udata,
		// EndOfFile: (head.Flags & EV_EOF) == EV_EOF,
		Type: eventType,
	}, nil
}
