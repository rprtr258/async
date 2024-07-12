package imhttp

import "reflect"

type Promise[T any] struct {
	ch <-chan T
}

func NewInstant[T any](value T) Promise[T] {
	return NewAsync(func() T {
		return value
	})
}

func NewAsync[T any](get func() T) Promise[T] {
	ch := make(chan T)
	go func() { ch <- get() }()
	return Promise[T]{ch}
}

func Map[T, R any](p Promise[T], fn func(T) R) Promise[R] {
	return NewAsync(func() R {
		return fn(p.Await())
	})
}

func FlatMap[T, R any](p Promise[T], fn func(T) Promise[R]) Promise[R] {
	return fn(p.Await())
}

func (p Promise[T]) Await() T {
	return <-p.ch
}

func (p Promise[T]) TryAwait() (T, bool) {
	select {
	case t := <-p.ch:
		return t, true
	default:
		return *new(T), false
	}
}

// TODO: remove T result, instead just return index of promise,
// awaiting which would not block (HOW BLYAT)
func Select[T any](ps ...Promise[T]) (int, T) {
	cases := make([]reflect.SelectCase, len(ps))
	for i, p := range ps {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(p.ch),
		}
	}
	i, val, _ := reflect.Select(cases)
	return i, val.Interface().(T)
}

type Result[T any] struct {
	Value T
	Error error
}
