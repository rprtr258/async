package imhttp

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
