package result

type Result[T any] struct {
	value T
	err   error
}

func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, err: nil}
}

func Err[T any](err error) Result[T] {
	var zero T
	return Result[T]{value: zero, err: err}
}

func (r Result[T]) IsOk() bool {
	return r.err == nil
}

func (r Result[T]) IsErr() bool {
	return r.err != nil
}

func (r Result[T]) Unwrap() T {
	if r.IsErr() {
		panic("called Unwrap on an Err value")
	}
	return r.value
}

func (r Result[T]) UnwrapErr() error {
	if r.IsOk() {
		panic("called UnwrapErr on an Ok value")
	}
	return r.err
}

func (r Result[T]) Expect(msg string) T {
	if r.IsErr() {
		panic(msg)
	}
	return r.value
}

func (r Result[T]) ExpectErr(msg string) error {
	if r.IsOk() {
		panic(msg)
	}
	return r.err
}

func (r Result[T]) GetOrElse(defaultValue T) T {
	if r.IsOk() {
		return r.value
	}
	return defaultValue
}

func (r Result[T]) GetOrElseFunc(defaultFunc func() T) T {
	if r.IsOk() {
		return r.value
	}
	return defaultFunc()
}

func (r Result[T]) Map(f func(T) interface{}) Result[interface{}] {
	if r.IsOk() {
		return Ok(f(r.value))
	}
	return Err[interface{}](r.err)
}

func (r Result[T]) MapErr(f func(error) error) Result[T] {
	if r.IsErr() {
		return Err[T](f(r.err))
	}
	return r
}

func (r Result[T]) AndThen(f func(T) Result[interface{}]) Result[interface{}] {
	if r.IsOk() {
		return f(r.value)
	}
	return Err[interface{}](r.err)
}

func (r Result[T]) Or(other Result[T]) Result[T] {
	if r.IsOk() {
		return r
	}
	return other
}

func (r Result[T]) And(other Result[T]) Result[T] {
	if r.IsErr() {
		return r
	}
	return other
}

func (r Result[T]) Filter(predicate func(T) bool, err error) Result[T] {
	if r.IsOk() && !predicate(r.value) {
		return Err[T](err)
	}
	return r
}

func Map[T, U any](r Result[T], f func(T) U) Result[U] {
	if r.IsOk() {
		return Ok(f(r.value))
	}
	return Err[U](r.err)
}

func AndThen[T, U any](r Result[T], f func(T) Result[U]) Result[U] {
	if r.IsOk() {
		return f(r.value)
	}
	return Err[U](r.err)
}
