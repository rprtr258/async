package imhttp

import "reflect"

type Future[T any] struct {
	ch <-chan T
}

func NewInstant[T any](value T) Future[T] {
	return NewAsync(func() T {
		return value
	})
}

func NewAsync[T any](get func() T) Future[T] {
	ch := make(chan T, 1)
	go func() { ch <- get(); close(ch) }()
	return Future[T]{ch}
}

func Map[T, R any](p Future[T], fn func(T) R) Future[R] {
	return NewAsync(func() R {
		return fn(p.Await())
	})
}

func FlatMap[T, R any](p Future[T], fn func(T) Future[R]) Future[R] {
	return fn(p.Await())
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

// TODO: remove T result, instead just return index of promise,
// awaiting which would not block (HOW BLYAT)
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

func Join[T any](fs ...Future[T]) []T {
	res := make([]T, len(fs))
	for i, f := range fs {
		res[i] = f.Await()
	}
	return res
}
