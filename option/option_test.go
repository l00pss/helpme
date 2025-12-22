package option

import (
	"fmt"
	"strconv"
	"testing"
)

func TestSome(t *testing.T) {
	opt := Some(42)
	if !opt.IsSome() {
		t.Error("Some(42) should be Some")
	}
	if opt.IsNone() {
		t.Error("Some(42) should not be None")
	}
	if opt.Unwrap() != 42 {
		t.Errorf("Some(42).Unwrap() = %v, want 42", opt.Unwrap())
	}
}

func TestNone(t *testing.T) {
	opt := None[int]()
	if opt.IsSome() {
		t.Error("None should not be Some")
	}
	if !opt.IsNone() {
		t.Error("None should be None")
	}
}

func TestUnwrapPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("None.Unwrap() should panic")
		}
	}()
	None[int]().Unwrap()
}

func TestExpect(t *testing.T) {
	opt := Some("hello")
	if opt.Expect("should have value") != "hello" {
		t.Error("Some.Expect should return value")
	}
}

func TestExpectPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("None.Expect() should panic")
		}
	}()
	None[string]().Expect("custom panic message")
}

func TestGetOrElse(t *testing.T) {
	someOpt := Some(42)
	if someOpt.GetOrElse(0) != 42 {
		t.Error("Some.GetOrElse should return wrapped value")
	}

	noneOpt := None[int]()
	if noneOpt.GetOrElse(99) != 99 {
		t.Error("None.GetOrElse should return default value")
	}
}

func TestGetOrElseFunc(t *testing.T) {
	someOpt := Some(42)
	if someOpt.GetOrElseFunc(func() int { return 0 }) != 42 {
		t.Error("Some.GetOrElseFunc should return wrapped value")
	}

	noneOpt := None[int]()
	if noneOpt.GetOrElseFunc(func() int { return 99 }) != 99 {
		t.Error("None.GetOrElseFunc should return function result")
	}
}

func TestMap(t *testing.T) {
	someOpt := Some(42)
	mapped := someOpt.Map(func(x int) interface{} { return x * 2 })
	if !mapped.IsSome() {
		t.Error("Map on Some should return Some")
	}
	if mapped.Unwrap() != 84 {
		t.Errorf("Map result = %v, want 84", mapped.Unwrap())
	}

	noneOpt := None[int]()
	mappedNone := noneOpt.Map(func(x int) interface{} { return x * 2 })
	if !mappedNone.IsNone() {
		t.Error("Map on None should return None")
	}
}

func TestGenericMap(t *testing.T) {
	someOpt := Some(42)
	mapped := Map(someOpt, func(x int) string { return strconv.Itoa(x) })
	if !mapped.IsSome() {
		t.Error("Generic Map on Some should return Some")
	}
	if mapped.Unwrap() != "42" {
		t.Errorf("Generic Map result = %v, want '42'", mapped.Unwrap())
	}

	noneOpt := None[int]()
	mappedNone := Map(noneOpt, func(x int) string { return strconv.Itoa(x) })
	if !mappedNone.IsNone() {
		t.Error("Generic Map on None should return None")
	}
}

func TestAndThen(t *testing.T) {
	someOpt := Some(42)
	result := someOpt.AndThen(func(x int) Option[interface{}] {
		if x > 0 {
			return Some(interface{}(x * 2))
		}
		return None[interface{}]()
	})
	if !result.IsSome() {
		t.Error("AndThen on Some should return Some")
	}
	if result.Unwrap() != 84 {
		t.Errorf("AndThen result = %v, want 84", result.Unwrap())
	}

	noneOpt := None[int]()
	resultNone := noneOpt.AndThen(func(x int) Option[interface{}] {
		return Some(interface{}(x * 2))
	})
	if !resultNone.IsNone() {
		t.Error("AndThen on None should return None")
	}
}

func TestGenericAndThen(t *testing.T) {
	someOpt := Some(42)
	result := AndThen(someOpt, func(x int) Option[string] {
		if x > 0 {
			return Some(strconv.Itoa(x))
		}
		return None[string]()
	})
	if !result.IsSome() {
		t.Error("Generic AndThen on Some should return Some")
	}
	if result.Unwrap() != "42" {
		t.Errorf("Generic AndThen result = %v, want '42'", result.Unwrap())
	}
}

func TestOr(t *testing.T) {
	someOpt := Some(42)
	other := Some(99)
	result := someOpt.Or(other)
	if result.Unwrap() != 42 {
		t.Error("Some.Or should return first Some")
	}

	noneOpt := None[int]()
	result2 := noneOpt.Or(other)
	if result2.Unwrap() != 99 {
		t.Error("None.Or should return second option")
	}
}

func TestAnd(t *testing.T) {
	someOpt := Some(42)
	other := Some(99)
	result := someOpt.And(other)
	if result.Unwrap() != 99 {
		t.Error("Some.And(Some) should return second option")
	}

	noneOpt := None[int]()
	result2 := noneOpt.And(other)
	if !result2.IsNone() {
		t.Error("None.And should return None")
	}
}

func TestFilter(t *testing.T) {
	someOpt := Some(42)
	filtered := someOpt.Filter(func(x int) bool { return x > 40 })
	if !filtered.IsSome() {
		t.Error("Filter with true predicate should return Some")
	}

	filtered2 := someOpt.Filter(func(x int) bool { return x < 40 })
	if !filtered2.IsNone() {
		t.Error("Filter with false predicate should return None")
	}

	noneOpt := None[int]()
	filtered3 := noneOpt.Filter(func(x int) bool { return true })
	if !filtered3.IsNone() {
		t.Error("Filter on None should return None")
	}
}

func TestContains(t *testing.T) {
	someOpt := Some(42)
	eq := func(a, b int) bool { return a == b }

	if !someOpt.Contains(42, eq) {
		t.Error("Some(42) should contain 42")
	}
	if someOpt.Contains(99, eq) {
		t.Error("Some(42) should not contain 99")
	}

	noneOpt := None[int]()
	if noneOpt.Contains(42, eq) {
		t.Error("None should not contain any value")
	}
}

func TestExists(t *testing.T) {
	someOpt := Some(42)
	if !someOpt.Exists(func(x int) bool { return x > 40 }) {
		t.Error("Some(42) should satisfy predicate x > 40")
	}
	if someOpt.Exists(func(x int) bool { return x < 40 }) {
		t.Error("Some(42) should not satisfy predicate x < 40")
	}

	noneOpt := None[int]()
	if noneOpt.Exists(func(x int) bool { return true }) {
		t.Error("None should not satisfy any predicate")
	}
}

func TestForAll(t *testing.T) {
	someOpt := Some(42)
	if !someOpt.ForAll(func(x int) bool { return x > 40 }) {
		t.Error("Some(42) should satisfy predicate x > 40")
	}
	if someOpt.ForAll(func(x int) bool { return x < 40 }) {
		t.Error("Some(42) should not satisfy predicate x < 40")
	}

	noneOpt := None[int]()
	if !noneOpt.ForAll(func(x int) bool { return false }) {
		t.Error("None should satisfy all predicates (vacuously true)")
	}
}

func TestToSlice(t *testing.T) {
	someOpt := Some(42)
	slice := someOpt.ToSlice()
	if len(slice) != 1 || slice[0] != 42 {
		t.Errorf("Some(42).ToSlice() = %v, want [42]", slice)
	}

	noneOpt := None[int]()
	sliceNone := noneOpt.ToSlice()
	if len(sliceNone) != 0 {
		t.Errorf("None.ToSlice() = %v, want []", sliceNone)
	}
}

func TestString(t *testing.T) {
	someOpt := Some(42)
	str := someOpt.String()
	expected := "Some(42)"
	if str != expected {
		t.Errorf("Some(42).String() = %v, want %v", str, expected)
	}

	noneOpt := None[int]()
	strNone := noneOpt.String()
	expectedNone := "None"
	if strNone != expectedNone {
		t.Errorf("None.String() = %v, want %v", strNone, expectedNone)
	}
}

func TestReplace(t *testing.T) {
	someOpt := Some(42)
	replaced := someOpt.Replace(99)
	if !replaced.IsSome() || replaced.Unwrap() != 99 {
		t.Error("Replace on Some should return Some with new value")
	}

	noneOpt := None[int]()
	replacedNone := noneOpt.Replace(99)
	if !replacedNone.IsNone() {
		t.Error("Replace on None should return None")
	}
}

func TestTake(t *testing.T) {
	someOpt := Some(42)
	taken := someOpt.Take()
	if !taken.IsSome() || taken.Unwrap() != 42 {
		t.Error("Take on Some should return copy of the option")
	}

	noneOpt := None[int]()
	takenNone := noneOpt.Take()
	if !takenNone.IsNone() {
		t.Error("Take on None should return None")
	}
}

func TestChaining(t *testing.T) {
	result := Some(5).
		Filter(func(x int) bool { return x > 0 }).
		Map(func(x int) interface{} { return x * 2 }).
		AndThen(func(x interface{}) Option[interface{}] {
			if val, ok := x.(int); ok && val > 5 {
				return Some(interface{}(fmt.Sprintf("Result: %d", val)))
			}
			return None[interface{}]()
		})

	if !result.IsSome() {
		t.Error("Chained operations should result in Some")
	}

	expected := "Result: 10"
	if result.Unwrap() != expected {
		t.Errorf("Chained result = %v, want %v", result.Unwrap(), expected)
	}
}

func BenchmarkSomeCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Some(i)
	}
}

func BenchmarkNoneCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = None[int]()
	}
}

func BenchmarkMapChain(b *testing.B) {
	opt := Some(42)
	for i := 0; i < b.N; i++ {
		_ = Map(Map(opt, func(x int) int { return x * 2 }), func(x int) string { return strconv.Itoa(x) })
	}
}
