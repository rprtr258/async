package imhttp

type Result[T any] struct {
	Value T
	Error error
}

func NewResult[T any](value T, err error) Result[T] {
	return Result[T]{value, err}
}

func Ok[T any](value T) Result[T] {
	return Result[T]{Value: value}
}

func Err[T any](err error) Result[T] {
	return Result[T]{Error: err}
}

func (r Result[T]) Unwrap() T {
	if r.Error != nil {
		panic(r.Error)
	}
	return r.Value
}
