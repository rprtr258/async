package imhttp

import "reflect"

type Promise[T any] struct {
	ch <-chan T
}

func NewAsync[T any](get func() T) Promise[T] {
	ch := make(chan T)
	go func() { ch <- get() }()
	return Promise[T]{ch}
}

func (p Promise[T]) Await() T {
	return <-p.ch
}

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
