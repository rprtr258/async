package imhttp

import "slices"

// actually is just a channel
type Stream[T any] struct {
	ch <-chan T
}

func NewStream[T any](ch <-chan T) Stream[T] {
	return Stream[T]{ch}
}

func NewGenerator[T any](next func() (T, bool)) Stream[T] {
	ch := make(chan T)
	go func() {
		for {
			x, ok := next()
			if !ok {
				close(ch)
				return
			}
			ch <- x
		}
	}()
	return Stream[T]{ch}
}

func NewInfinite[T any](next func() T) Stream[T] {
	ch := make(chan T)
	go func() {
		for {
			ch <- next()
		}
	}()
	return Stream[T]{ch}
}

func NewSelectAll[T any](ss ...Stream[T]) Stream[T] {
	ch := make(chan T)
	go func() {
		futures := make([]Future[Option[T]], len(ss))
		for i, s := range ss {
			futures[i] = s.Next()
		}

		for { // loop, yielding values
			if len(ss) == 0 {
				break
			}

		RETRY: // loop yielding single value among closed streams
			i, f := Select(futures...)
			if !f.Valid {
				// stream closed, remove it and retry
				ss = slices.Delete(ss, i, i+1)[:len(ss)-1]
				futures = slices.Delete(futures, i, i+1)[:len(futures)-1]
				goto RETRY
			}

			// got value, produce it and get new one in place of old one
			ch <- f.Value
			futures[i] = ss[i].Next()
		}
		close(ch)
	}()
	return NewStream(ch)
}

type Option[T any] struct {
	Value T
	Valid bool
}

func (s Stream[T]) Next() Future[Option[T]] {
	return NewFuture(func() Option[T] {
		value, ok := <-s.ch
		return Option[T]{value, ok}
	})
}
