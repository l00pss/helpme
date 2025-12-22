package option

import "fmt"

type Option[T any] struct {
	value *T
}

func Some[T any](value T) Option[T] {
	return Option[T]{value: &value}
}

func None[T any]() Option[T] {
	return Option[T]{value: nil}
}

func (o Option[T]) IsSome() bool {
	return o.value != nil
}

func (o Option[T]) IsNone() bool {
	return o.value == nil
}

func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic("called Unwrap on a None value")
	}
	return *o.value
}

func (o Option[T]) Expect(msg string) T {
	if o.IsNone() {
		panic(msg)
	}
	return *o.value
}

func (o Option[T]) GetOrElse(defaultValue T) T {
	if o.IsSome() {
		return *o.value
	}
	return defaultValue
}

func (o Option[T]) GetOrElseFunc(defaultFunc func() T) T {
	if o.IsSome() {
		return *o.value
	}
	return defaultFunc()
}

func (o Option[T]) Map(f func(T) interface{}) Option[interface{}] {
	if o.IsSome() {
		return Some(f(*o.value))
	}
	return None[interface{}]()
}

func (o Option[T]) AndThen(f func(T) Option[interface{}]) Option[interface{}] {
	if o.IsSome() {
		return f(*o.value)
	}
	return None[interface{}]()
}

func (o Option[T]) Or(other Option[T]) Option[T] {
	if o.IsSome() {
		return o
	}
	return other
}

func (o Option[T]) And(other Option[T]) Option[T] {
	if o.IsNone() {
		return o
	}
	return other
}

func (o Option[T]) Filter(predicate func(T) bool) Option[T] {
	if o.IsSome() && predicate(*o.value) {
		return o
	}
	return None[T]()
}

func (o Option[T]) Contains(value T, eq func(T, T) bool) bool {
	if o.IsNone() {
		return false
	}
	return eq(*o.value, value)
}

func (o Option[T]) Exists(predicate func(T) bool) bool {
	return o.IsSome() && predicate(*o.value)
}

func (o Option[T]) ForAll(predicate func(T) bool) bool {
	return o.IsNone() || predicate(*o.value)
}

func (o Option[T]) ToSlice() []T {
	if o.IsSome() {
		return []T{*o.value}
	}
	return []T{}
}

func (o Option[T]) String() string {
	if o.IsSome() {
		return "Some(" + fmt.Sprintf("%v", *o.value) + ")"
	}
	return "None"
}

func (o Option[T]) Flatten() Option[T] {
	if o.IsSome() {
		return o
	}
	return None[T]()
}

func (o Option[T]) Replace(value T) Option[T] {
	if o.IsSome() {
		return Some(value)
	}
	return None[T]()
}

func (o Option[T]) Take() Option[T] {
	if o.IsSome() {
		return o
	}
	return None[T]()
}

func Map[T, U any](o Option[T], f func(T) U) Option[U] {
	if o.IsSome() {
		return Some(f(*o.value))
	}
	return None[U]()
}

func AndThen[T, U any](o Option[T], f func(T) Option[U]) Option[U] {
	if o.IsSome() {
		return f(*o.value)
	}
	return None[U]()
}
