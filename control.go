package imhttp

import "fmt"

func While(cond func() bool) Promise[bool] {
	return NewAsync(cond)
}

func exampleWhile() {
	n := 0
	for it := While(func() bool { return n < 10 }); it.Await(); {
		fmt.Println(n)
		n++
	}
}

type maybe[T any] struct {
	v T
	b bool
}

func (m maybe[T]) Unpack() (T, bool) {
	return m.v, m.b
}

func While2[T any](cond func() bool, get func() T) Promise[maybe[T]] {
	return NewAsync(func() maybe[T] {
		if cond() {
			return maybe[T]{get(), true}
		}
		return maybe[T]{*new(T), false}
	})
}

func exampleWhile2() {
	n := 0
	for it := While2(func() bool { return n < 10 }, func() int { return n }); ; {
		v, ok := it.Await().Unpack()
		if !ok {
			break
		}

		fmt.Println(v)
		n++
	}
}

func If(cond bool) Promise[bool] {
	first := true
	return NewAsync(func() bool {
		if first {
			first = false
			return cond
		}
		return false
	})
}

func exampleIf() {
	n := 0
	for it := If(n < 10); it.Await(); {
		fmt.Println(n)
		n++
	}
}
