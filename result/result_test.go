package result

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

func TestOk(t *testing.T) {
	result := Ok(42)
	if !result.IsOk() {
		t.Error("Ok(42) should be Ok")
	}
	if result.IsErr() {
		t.Error("Ok(42) should not be Err")
	}
	if result.Unwrap() != 42 {
		t.Errorf("Ok(42).Unwrap() = %v, want 42", result.Unwrap())
	}
}

func TestErr(t *testing.T) {
	err := errors.New("test error")
	result := Err[int](err)
	if result.IsOk() {
		t.Error("Err should not be Ok")
	}
	if !result.IsErr() {
		t.Error("Err should be Err")
	}
	if result.UnwrapErr() != err {
		t.Errorf("Err.UnwrapErr() = %v, want %v", result.UnwrapErr(), err)
	}
}

func TestUnwrapPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Err.Unwrap() should panic")
		}
	}()
	Err[int](errors.New("test")).Unwrap()
}

func TestUnwrapErrPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Ok.UnwrapErr() should panic")
		}
	}()
	Ok(42).UnwrapErr()
}

func TestExpect(t *testing.T) {
	result := Ok("hello")
	if result.Expect("should have value") != "hello" {
		t.Error("Ok.Expect should return value")
	}
}

func TestExpectPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Err.Expect() should panic")
		}
	}()
	Err[string](errors.New("test")).Expect("custom panic message")
}

func TestExpectErr(t *testing.T) {
	err := errors.New("test error")
	result := Err[int](err)
	if result.ExpectErr("should have error") != err {
		t.Error("Err.ExpectErr should return error")
	}
}

func TestExpectErrPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Ok.ExpectErr() should panic")
		}
	}()
	Ok(42).ExpectErr("custom panic message")
}

func TestGetOrElse(t *testing.T) {
	okResult := Ok(42)
	if okResult.GetOrElse(0) != 42 {
		t.Error("Ok.GetOrElse should return wrapped value")
	}

	errResult := Err[int](errors.New("test"))
	if errResult.GetOrElse(99) != 99 {
		t.Error("Err.GetOrElse should return default value")
	}
}

func TestGetOrElseFunc(t *testing.T) {
	okResult := Ok(42)
	if okResult.GetOrElseFunc(func() int { return 0 }) != 42 {
		t.Error("Ok.GetOrElseFunc should return wrapped value")
	}

	errResult := Err[int](errors.New("test"))
	if errResult.GetOrElseFunc(func() int { return 99 }) != 99 {
		t.Error("Err.GetOrElseFunc should return function result")
	}
}

func TestMap(t *testing.T) {
	okResult := Ok(42)
	mapped := okResult.Map(func(x int) interface{} { return x * 2 })
	if !mapped.IsOk() {
		t.Error("Map on Ok should return Ok")
	}
	if mapped.Unwrap() != 84 {
		t.Errorf("Map result = %v, want 84", mapped.Unwrap())
	}

	errResult := Err[int](errors.New("test"))
	mappedErr := errResult.Map(func(x int) interface{} { return x * 2 })
	if !mappedErr.IsErr() {
		t.Error("Map on Err should return Err")
	}
}

func TestGenericMap(t *testing.T) {
	okResult := Ok(42)
	mapped := Map(okResult, func(x int) string { return strconv.Itoa(x) })
	if !mapped.IsOk() {
		t.Error("Generic Map on Ok should return Ok")
	}
	if mapped.Unwrap() != "42" {
		t.Errorf("Generic Map result = %v, want '42'", mapped.Unwrap())
	}

	errResult := Err[int](errors.New("test"))
	mappedErr := Map(errResult, func(x int) string { return strconv.Itoa(x) })
	if !mappedErr.IsErr() {
		t.Error("Generic Map on Err should return Err")
	}
}

func TestMapErr(t *testing.T) {
	okResult := Ok(42)
	mappedOk := okResult.MapErr(func(e error) error { return errors.New("new error") })
	if !mappedOk.IsOk() || mappedOk.Unwrap() != 42 {
		t.Error("MapErr on Ok should return unchanged Ok")
	}

	errResult := Err[int](errors.New("original"))
	mappedErr := errResult.MapErr(func(e error) error { return errors.New("mapped error") })
	if !mappedErr.IsErr() {
		t.Error("MapErr on Err should return Err")
	}
	if mappedErr.UnwrapErr().Error() != "mapped error" {
		t.Errorf("MapErr result = %v, want 'mapped error'", mappedErr.UnwrapErr().Error())
	}
}

func TestAndThen(t *testing.T) {
	okResult := Ok(42)
	result := okResult.AndThen(func(x int) Result[interface{}] {
		if x > 0 {
			return Ok(interface{}(x * 2))
		}
		return Err[interface{}](errors.New("negative"))
	})
	if !result.IsOk() {
		t.Error("AndThen on Ok should return Ok")
	}
	if result.Unwrap() != 84 {
		t.Errorf("AndThen result = %v, want 84", result.Unwrap())
	}

	errResult := Err[int](errors.New("test"))
	resultErr := errResult.AndThen(func(x int) Result[interface{}] {
		return Ok(interface{}(x * 2))
	})
	if !resultErr.IsErr() {
		t.Error("AndThen on Err should return Err")
	}
}

func TestGenericAndThen(t *testing.T) {
	okResult := Ok(42)
	result := AndThen(okResult, func(x int) Result[string] {
		if x > 0 {
			return Ok(strconv.Itoa(x))
		}
		return Err[string](errors.New("negative"))
	})
	if !result.IsOk() {
		t.Error("Generic AndThen on Ok should return Ok")
	}
	if result.Unwrap() != "42" {
		t.Errorf("Generic AndThen result = %v, want '42'", result.Unwrap())
	}
}

func TestOr(t *testing.T) {
	okResult := Ok(42)
	other := Ok(99)
	result := okResult.Or(other)
	if result.Unwrap() != 42 {
		t.Error("Ok.Or should return first Ok")
	}

	errResult := Err[int](errors.New("test"))
	result2 := errResult.Or(other)
	if result2.Unwrap() != 99 {
		t.Error("Err.Or should return second result")
	}
}

func TestAnd(t *testing.T) {
	okResult := Ok(42)
	other := Ok(99)
	result := okResult.And(other)
	if result.Unwrap() != 99 {
		t.Error("Ok.And(Ok) should return second result")
	}

	errResult := Err[int](errors.New("test"))
	result2 := errResult.And(other)
	if !result2.IsErr() {
		t.Error("Err.And should return Err")
	}
}

func TestFilter(t *testing.T) {
	okResult := Ok(42)
	testErr := errors.New("filter failed")

	filtered := okResult.Filter(func(x int) bool { return x > 40 }, testErr)
	if !filtered.IsOk() {
		t.Error("Filter with true predicate should return Ok")
	}

	filtered2 := okResult.Filter(func(x int) bool { return x < 40 }, testErr)
	if !filtered2.IsErr() {
		t.Error("Filter with false predicate should return Err")
	}
	if filtered2.UnwrapErr() != testErr {
		t.Error("Filter should return provided error")
	}

	errResult := Err[int](errors.New("original"))
	filtered3 := errResult.Filter(func(x int) bool { return true }, testErr)
	if !filtered3.IsErr() {
		t.Error("Filter on Err should return Err")
	}
	if filtered3.UnwrapErr().Error() != "original" {
		t.Error("Filter on Err should preserve original error")
	}
}

func TestChaining(t *testing.T) {
	result := Ok(5).
		Filter(func(x int) bool { return x > 0 }, errors.New("not positive")).
		Map(func(x int) interface{} { return x * 2 }).
		AndThen(func(x interface{}) Result[interface{}] {
			if val, ok := x.(int); ok && val > 5 {
				return Ok(interface{}(fmt.Sprintf("Result: %d", val)))
			}
			return Err[interface{}](errors.New("value too small"))
		})

	if !result.IsOk() {
		t.Error("Chained operations should result in Ok")
	}

	expected := "Result: 10"
	if result.Unwrap() != expected {
		t.Errorf("Chained result = %v, want %v", result.Unwrap(), expected)
	}
}

func TestErrorChaining(t *testing.T) {
	result := Err[int](errors.New("initial error")).
		Map(func(x int) interface{} { return x * 2 }).
		AndThen(func(x interface{}) Result[interface{}] {
			return Ok(interface{}("should not reach here"))
		})

	if !result.IsErr() {
		t.Error("Error should propagate through chain")
	}
	if result.UnwrapErr().Error() != "initial error" {
		t.Error("Original error should be preserved")
	}
}

func TestFilterFailureChain(t *testing.T) {
	result := Ok(5).
		Filter(func(x int) bool { return x > 10 }, errors.New("too small")).
		Map(func(x int) interface{} { return x * 2 })

	if !result.IsErr() {
		t.Error("Filter failure should stop chain")
	}
	if result.UnwrapErr().Error() != "too small" {
		t.Error("Filter error should be preserved")
	}
}

// Test different value types
func TestStringResult(t *testing.T) {
	result := Ok("hello")
	mapped := Map(result, func(s string) int { return len(s) })

	if !mapped.IsOk() {
		t.Error("String to int mapping should work")
	}
	if mapped.Unwrap() != 5 {
		t.Errorf("String length = %v, want 5", mapped.Unwrap())
	}
}

func TestStructResult(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	person := Person{Name: "John", Age: 30}
	result := Ok(person)

	mapped := Map(result, func(p Person) string { return p.Name })
	if !mapped.IsOk() {
		t.Error("Struct mapping should work")
	}
	if mapped.Unwrap() != "John" {
		t.Errorf("Person name = %v, want 'John'", mapped.Unwrap())
	}
}

// Benchmark tests
func BenchmarkOkCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Ok(i)
	}
}

func BenchmarkErrCreation(b *testing.B) {
	err := errors.New("test error")
	for i := 0; i < b.N; i++ {
		_ = Err[int](err)
	}
}

func BenchmarkMapChain(b *testing.B) {
	result := Ok(42)
	for i := 0; i < b.N; i++ {
		_ = Map(Map(result, func(x int) int { return x * 2 }), func(x int) string { return strconv.Itoa(x) })
	}
}

func BenchmarkAndThenChain(b *testing.B) {
	result := Ok(42)
	for i := 0; i < b.N; i++ {
		_ = AndThen(result, func(x int) Result[string] {
			return Ok(strconv.Itoa(x * 2))
		})
	}
}

// Error handling patterns
func TestDivideByZero(t *testing.T) {
	divide := func(a, b int) Result[int] {
		if b == 0 {
			return Err[int](errors.New("division by zero"))
		}
		return Ok(a / b)
	}

	result1 := divide(10, 2)
	if !result1.IsOk() || result1.Unwrap() != 5 {
		t.Error("10/2 should equal 5")
	}

	result2 := divide(10, 0)
	if !result2.IsErr() {
		t.Error("10/0 should be an error")
	}
	if result2.UnwrapErr().Error() != "division by zero" {
		t.Error("Should have division by zero error")
	}
}

func TestParseInt(t *testing.T) {
	parseInt := func(s string) Result[int] {
		val, err := strconv.Atoi(s)
		if err != nil {
			return Err[int](err)
		}
		return Ok(val)
	}

	result1 := parseInt("42")
	if !result1.IsOk() || result1.Unwrap() != 42 {
		t.Error("'42' should parse to 42")
	}

	result2 := parseInt("not a number")
	if !result2.IsErr() {
		t.Error("'not a number' should be an error")
	}
}
