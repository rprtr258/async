package imhttp

import "reflect"

// actually is just 1-buffered channel which should be received once
type Future[T any] struct {
	ch <-chan T
}

func NewReady[T any](value T) Future[T] {
	return NewFuture(func() T {
		return value
	})
}

func NewFuture[T any](get func() T) Future[T] {
	ch := make(chan T, 1)
	go func() { ch <- get(); close(ch) }()
	return Future[T]{ch}
}

func Map[T, R any](p Future[T], fn func(T) R) Future[R] {
	return NewFuture(func() R {
		return fn(p.Await())
	})
}

func FlatMap[T, R any](p Future[T], fn func(T) Future[R]) Future[R] {
	return fn(p.Await())
}

func (p Future[T]) Raw() <-chan T {
	return p.ch
}

func (p Future[T]) Await() T {
	return <-p.ch
}

func (p Future[T]) TryAwait() (T, bool) {
	select {
	case t := <-p.ch:
		return t, true
	default:
		return *new(T), false
	}
}

func Join[T any](fs ...Future[T]) []T {
	res := make([]T, len(fs))
	for i, f := range fs {
		res[i] = f.Await()
	}
	return res
}

func Flatten[T any](fs ...Future[T]) Future[[]T] {
	return NewFuture(func() []T {
		return Join(fs...)
	})
}

// TODO: remove T result, instead just return index of promise,
// awaiting which would not block (HOW BLYAT)
// TODO: how to use correctly?
func Select[T any](fs ...Future[T]) (int, T) {
	cases := make([]reflect.SelectCase, len(fs))
	for i, f := range fs {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(f.ch),
		}
	}
	i, val, _ := reflect.Select(cases)
	return i, val.Interface().(T)
}

// TODO: how to use correctly?
func Select2[A, B any](
	fa Future[A], funa func(A),
	fb Future[B], funb func(B),
) {
	i, val, _ := reflect.Select([]reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(fa.ch),
		},
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(fb.ch),
		},
	})
	switch i {
	case 0:
		funa(val.Interface().(A))
	case 1:
		funb(val.Interface().(B))
	}
}

// TODO: how to use correctly?
func Select2D[A, B any](
	fa Future[A], funa func(A),
	fb Future[B], funb func(B),
	fundef func(),
) {
	i, val, _ := reflect.Select([]reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(fa.ch),
		},
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(fb.ch),
		},
		{
			Dir: reflect.SelectDefault,
		},
	})
	switch i {
	case 0:
		funa(val.Interface().(A))
	case 1:
		funb(val.Interface().(B))
	case 2:
		fundef()
	}
}
