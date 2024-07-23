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

func (o Option[T]) Unwrap() T {
	if !o.Valid {
		panic("value is not valid")
	}
	return o.Value
}

func (o Option[T]) Unpack() (T, bool) {
	return o.Value, o.Valid
}

func (s Stream[T]) Next() Future[Option[T]] {
	return NewFuture(func() Option[T] {
		value, ok := <-s.ch
		return Option[T]{value, ok}
	})
}

func (s Stream[T]) ForEachConcurrent(fn func(T) Future[struct{}]) Future[struct{}] {
	return NewFuture(func() struct{} {
		set := NewFutureSet[struct{}]()
		var next func() Future[struct{}]
		next = func() Future[struct{}] {
			return s.Next().Then(func(o Option[T]) {
				if o.Valid {
					set.Push(fn(o.Value))
					set.Push(next())
				}
			})
		}
		set.Push(next())
		set.IntoIter()(func(f Future[struct{}]) {
			f.Await()
		})
		return struct{}{}
	})
}

func (s Stream[T]) Then(fn func(T) Future[struct{}]) Stream[struct{}] {
	return NewGenerator(func() (struct{}, bool) {
		o := s.Next().Await()
		if !o.Valid {
			return struct{}{}, false
		}
		return fn(o.Value).Await(), true
	})
}
