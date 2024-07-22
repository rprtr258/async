package imhttp

import (
	"slices"

	"golang.org/x/exp/maps"
)

type FutureSet[T any] struct {
	data map[Future[T]]struct{}
}

func NewFutureSet[T any]() FutureSet[T] {
	return FutureSet[T]{map[Future[T]]struct{}{}}
}

func (fs *FutureSet[T]) Len() int {
	return len(fs.data)
}

func (fs *FutureSet[T]) Push(f Future[T]) {
	fs.data[f] = struct{}{}
}

func (fs *FutureSet[T]) Iter() func(func(Future[T])) {
	return func(yield func(Future[T])) {
		for f := range fs.data {
			yield(f)
		}
	}
}

func (fs *FutureSet[T]) Stream() Stream[T] {
	ch := make(chan T)
	go func() {
		for {
			futures := maps.Keys(fs.data)
			i, value := Select(futures...)
			delete(fs.data, futures[i])
			ch <- value
		}
	}()
	return NewStream(ch)
}

func (fs *FutureSet[T]) Clear() {
	clear(fs.data)
}

type FutureSetOrdered[T any] struct {
	data []Future[T]
}

func NewFutureSetOrdered[T any]() FutureSetOrdered[T] {
	return FutureSetOrdered[T]{nil}
}

func (fs *FutureSetOrdered[T]) Len() int {
	return len(fs.data)
}

func (fs *FutureSetOrdered[T]) Push(f Future[T]) {
	fs.data = append(fs.data, f)
}

func (fs *FutureSetOrdered[T]) Iter() func(func(Future[T])) {
	return func(yield func(Future[T])) {
		for _, f := range fs.data {
			yield(f)
		}
	}
}

func (fs *FutureSetOrdered[T]) Clear() {
	fs.data = nil
}

func (fs *FutureSetOrdered[T]) Stream() Stream[T] {
	ch := make(chan T)
	go func() {
		for {
			i, value := Select(fs.data...)
			fs.data = slices.Delete(fs.data, i, i+1)[:len(fs.data)-1]
			ch <- value
		}
	}()
	return NewStream(ch)
}
