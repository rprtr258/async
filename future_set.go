package imhttp

import (
	"slices"
)

type FutureSet[T any] struct {
	data []Future[T]
}

func NewFutureSet[T any]() FutureSet[T] {
	return FutureSet[T]{nil}
}

func (fs *FutureSet[T]) Len() int {
	return len(fs.data)
}

func (fs *FutureSet[T]) Push(f Future[T]) {
	fs.data = append(fs.data, f)
}

func (fs *FutureSet[T]) PushStream(s Stream[T]) {
	var next func() Future[T]
	next = func() Future[T] {
		return NewFuture(func() T {
			req, ok := s.Next().Await().Unpack()
			if ok {
				fs.Push(NewReady(req))
				fs.Push(next())
			}
			return req
		})
	}
	fs.Push(next())
}

func (fs *FutureSet[T]) IntoIter() func(func(Future[T])) {
	return func(yield func(Future[T])) {
		for len(fs.data) > 0 {
			n := len(fs.data)
			for _, f := range fs.data[:n] {
				yield(f)
			}
			fs.data = fs.data[n:]
		}
	}
}

func (fs *FutureSet[T]) Stream() Stream[T] {
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

func (fs *FutureSet[T]) Clear() {
	clear(fs.data)
}
