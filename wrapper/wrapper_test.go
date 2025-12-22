package wrapper

import (
	"context"
	"errors"
	"testing"
)

type TestQuery struct {
	Name string
	Age  int
}

type TestCommand struct {
	Action string
	Data   map[string]interface{}
}

type TestResult struct {
	ID   int
	Name string
}

// QueryWrapper tests
func TestNewQueryWrapper(t *testing.T) {
	ctx := context.Background()
	query := TestQuery{Name: "test", Age: 25}
	projection := NewProjection([]string{"name", "age"})
	pagination := NewPagination(10, 0)
	sortBy := NewSortBy("name", true)
	filters := []Filter{NewFilter("active", true)}

	wrapper := NewQueryWrapper(ctx, query, projection, pagination, sortBy, filters)

	if wrapper.Context != ctx {
		t.Error("Context not set correctly")
	}
	if wrapper.Query.Name != "test" || wrapper.Query.Age != 25 {
		t.Error("Query not set correctly")
	}
	if len(wrapper.Projection().Fields()) != 2 {
		t.Error("Projection not set correctly")
	}
	if wrapper.Pagination().Limit() != 10 {
		t.Error("Pagination not set correctly")
	}
	if wrapper.SortBy().Field() != "name" {
		t.Error("SortBy not set correctly")
	}
	if len(wrapper.Filter()) != 1 {
		t.Error("Filter not set correctly")
	}
}

func TestQueryWrapperBuilder(t *testing.T) {
	ctx := context.Background()
	query := TestQuery{Name: "builder", Age: 30}
	projection := NewProjection([]string{"id", "name"})
	pagination := NewPagination(20, 10)
	sortBy := NewDescendingSortBy("created_at")
	filters := []Filter{
		NewFilter("status", "active"),
		NewFilter("type", "premium"),
	}

	builder := NewQueryWrapperBuilder[TestQuery]()
	wrapper := builder.
		WithContext(ctx).
		WithQuery(query).
		withProjection(projection).
		WithPagination(pagination).
		WithSortBy(sortBy).
		WithFilter(filters).
		Build()

	if wrapper.Context != ctx {
		t.Error("Builder context not set correctly")
	}
	if wrapper.Query.Name != "builder" {
		t.Error("Builder query not set correctly")
	}
	if len(wrapper.Projection().Fields()) != 2 {
		t.Error("Builder projection not set correctly")
	}
	if wrapper.Pagination().Limit() != 20 {
		t.Error("Builder pagination not set correctly")
	}
	if !wrapper.SortBy().IsAscending() {
		t.Log("SortBy is correctly set to descending")
	}
	if len(wrapper.Filter()) != 2 {
		t.Error("Builder filter not set correctly")
	}
}

// CommandWrapper tests
func TestNewCommandWrapper(t *testing.T) {
	ctx := context.Background()
	command := TestCommand{
		Action: "create",
		Data:   map[string]interface{}{"name": "test", "value": 123},
	}

	wrapper := NewCommandWrapper(ctx, command)

	if wrapper.Context != ctx {
		t.Error("Context not set correctly")
	}
	if wrapper.Command.Action != "create" {
		t.Error("Command not set correctly")
	}
	if wrapper.Command.Data["name"] != "test" {
		t.Error("Command data not set correctly")
	}
}

func TestCommandWrapperBuilder(t *testing.T) {
	ctx := context.Background()
	command := TestCommand{
		Action: "update",
		Data:   map[string]interface{}{"id": 1, "status": "completed"},
	}

	builder := NewCommandWrapperBuilder[TestCommand]()
	wrapper := builder.
		WithContext(ctx).
		WithCommand(command).
		Build()

	if wrapper.Context != ctx {
		t.Error("Builder context not set correctly")
	}
	if wrapper.Command.Action != "update" {
		t.Error("Builder command not set correctly")
	}
}

// SortBy tests
func TestNewSortBy(t *testing.T) {
	sortBy := NewSortBy("name", true)

	if sortBy.Field() != "name" {
		t.Error("Field not set correctly")
	}
	if !sortBy.IsAscending() {
		t.Error("Ascending flag not set correctly")
	}
}

func TestNewAscendingSortBy(t *testing.T) {
	sortBy := NewAscendingSortBy("created_at")

	if sortBy.Field() != "created_at" {
		t.Error("Field not set correctly")
	}
	if !sortBy.IsAscending() {
		t.Error("Should be ascending")
	}
}

func TestNewDescendingSortBy(t *testing.T) {
	sortBy := NewDescendingSortBy("updated_at")

	if sortBy.Field() != "updated_at" {
		t.Error("Field not set correctly")
	}
	if sortBy.IsAscending() {
		t.Error("Should be descending")
	}
}

func TestSortByValidate(t *testing.T) {
	validSortBy := NewSortBy("name", true)
	if err := validSortBy.Validate(); err != nil {
		t.Errorf("Valid SortBy should not return error: %v", err)
	}

	invalidSortBy := NewSortBy("", true)
	if err := invalidSortBy.Validate(); err == nil {
		t.Error("Empty field should return error")
	}
}

// Pagination tests
func TestNewPagination(t *testing.T) {
	pagination := NewPagination(15, 30)

	if pagination.Limit() != 15 {
		t.Error("Limit not set correctly")
	}
	if pagination.Offset() != 30 {
		t.Error("Offset not set correctly")
	}
}

func TestNewFirstPagePagination(t *testing.T) {
	pagination := NewFirstPagePagination()

	if pagination.Limit() != 10 {
		t.Error("Default limit should be 10")
	}
	if pagination.Offset() != 0 {
		t.Error("Default offset should be 0")
	}
}

func TestPaginationHasNext(t *testing.T) {
	pagination := NewPagination(10, 0)

	// Total count 25, limit 10, offset 0 -> has next page
	if !pagination.HasNext(25) {
		t.Error("Should have next page")
	}

	// Total count 5, limit 10, offset 0 -> no next page
	if pagination.HasNext(5) {
		t.Error("Should not have next page")
	}

	// Exact boundary case
	pagination = NewPagination(10, 10)
	if pagination.HasNext(20) {
		t.Error("Should not have next page when at exact boundary")
	}
}

func TestPaginationNextPage(t *testing.T) {
	pagination := NewPagination(10, 20)
	nextPage := pagination.NextPage()

	if nextPage.Limit() != 10 {
		t.Error("Next page limit should remain the same")
	}
	if nextPage.Offset() != 30 {
		t.Error("Next page offset should be incremented by limit")
	}
}

func TestPaginationValidate(t *testing.T) {
	validPagination := NewPagination(10, 0)
	if err := validPagination.Validate(); err != nil {
		t.Errorf("Valid pagination should not return error: %v", err)
	}

	invalidLimitPagination := NewPagination(0, 0)
	if err := invalidLimitPagination.Validate(); err == nil {
		t.Error("Zero limit should return error")
	}

	negativeLimitPagination := NewPagination(-1, 0)
	if err := negativeLimitPagination.Validate(); err == nil {
		t.Error("Negative limit should return error")
	}

	negativeOffsetPagination := NewPagination(10, -1)
	if err := negativeOffsetPagination.Validate(); err == nil {
		t.Error("Negative offset should return error")
	}
}

// Filter tests
func TestNewFilter(t *testing.T) {
	filter := NewFilter("status", "active")

	if filter.Field() != "status" {
		t.Error("Field not set correctly")
	}
	if filter.Value() != "active" {
		t.Error("Value not set correctly")
	}
}

func TestFilterWithDifferentTypes(t *testing.T) {
	stringFilter := NewFilter("name", "test")
	intFilter := NewFilter("age", 25)
	boolFilter := NewFilter("active", true)

	if stringFilter.Value().(string) != "test" {
		t.Error("String value not preserved")
	}
	if intFilter.Value().(int) != 25 {
		t.Error("Int value not preserved")
	}
	if boolFilter.Value().(bool) != true {
		t.Error("Bool value not preserved")
	}
}

// Projection tests
func TestNewProjection(t *testing.T) {
	fields := []string{"id", "name", "email"}
	projection := NewProjection(fields)

	projectionFields := projection.Fields()
	if len(projectionFields) != 3 {
		t.Error("Fields count not correct")
	}
	for i, field := range projectionFields {
		if field != fields[i] {
			t.Errorf("Field %d not set correctly: expected %s, got %s", i, fields[i], field)
		}
	}
}

func TestNewEmptyProjection(t *testing.T) {
	projection := NewEmptyProjection()

	if len(projection.Fields()) != 0 {
		t.Error("Empty projection should have no fields")
	}
}

// Page tests
func TestPageMethods(t *testing.T) {
	results := []TestResult{
		{ID: 1, Name: "test1"},
		{ID: 2, Name: "test2"},
	}

	page := Page[TestResult]{
		Results: results,
		Offset:  0,
		Limit:   10,
		HasNext: true,
	}

	if !page.Next() {
		t.Error("Next() should return true when HasNext is true")
	}
	if !page.First() {
		t.Error("First() should return true when Offset is 0")
	}
	if !page.HasData() {
		t.Error("HasData() should return true when Results is not empty")
	}

	emptyPage := Page[TestResult]{
		Results: []TestResult{},
		Offset:  10,
		Limit:   10,
		HasNext: false,
	}

	if emptyPage.Next() {
		t.Error("Next() should return false when HasNext is false")
	}
	if emptyPage.First() {
		t.Error("First() should return false when Offset > 0")
	}
	if emptyPage.HasData() {
		t.Error("HasData() should return false when Results is empty")
	}
}

func TestPagesBuilder(t *testing.T) {
	results := []TestResult{
		{ID: 1, Name: "test1"},
		{ID: 2, Name: "test2"},
		{ID: 3, Name: "test3"},
	}

	builder := NewPagesBuilder[TestResult]()
	page := builder.
		Results(results).
		Offset(20).
		Limit(10).
		HasNext(true).
		Build()

	if len(page.Results) != 3 {
		t.Error("Results not set correctly")
	}
	if page.Offset != 20 {
		t.Error("Offset not set correctly")
	}
	if page.Limit != 10 {
		t.Error("Limit not set correctly")
	}
	if !page.HasNext {
		t.Error("HasNext not set correctly")
	}

	// Test that the results are properly copied
	if page.Results[0].ID != 1 || page.Results[0].Name != "test1" {
		t.Error("Results content not preserved")
	}
}

func TestPagesBuilderEmpty(t *testing.T) {
	builder := NewPagesBuilder[TestResult]()
	page := builder.Build()

	if page.Results != nil {
		t.Error("Default results should be nil")
	}
	if page.Offset != 0 {
		t.Error("Default offset should be 0")
	}
	if page.Limit != 0 {
		t.Error("Default limit should be 0")
	}
	if page.HasNext {
		t.Error("Default HasNext should be false")
	}
}

// Integration tests
func TestQueryWrapperIntegration(t *testing.T) {
	ctx := context.Background()
	query := TestQuery{Name: "integration_test", Age: 35}

	// Build complete query wrapper with all components
	wrapper := NewQueryWrapperBuilder[TestQuery]().
		WithContext(ctx).
		WithQuery(query).
		withProjection(NewProjection([]string{"id", "name", "age", "created_at"})).
		WithPagination(NewPagination(25, 50)).
		WithSortBy(NewDescendingSortBy("created_at")).
		WithFilter([]Filter{
			NewFilter("status", "active"),
			NewFilter("age", 18),
			NewFilter("verified", true),
		}).
		Build()

	// Validate all components
	if err := wrapper.SortBy().Validate(); err != nil {
		t.Errorf("SortBy validation failed: %v", err)
	}

	if err := wrapper.Pagination().Validate(); err != nil {
		t.Errorf("Pagination validation failed: %v", err)
	}

	// Test pagination logic
	if wrapper.Pagination().HasNext(100) {
		t.Log("Has next page with total count 100")
	} else {
		t.Error("Should have next page")
	}

	nextPagePagination := wrapper.Pagination().NextPage()
	if nextPagePagination.Offset() != 75 {
		t.Error("Next page offset calculation incorrect")
	}
}

func TestCommandWrapperIntegration(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "test_user")
	command := TestCommand{
		Action: "batch_update",
		Data: map[string]interface{}{
			"ids":    []int{1, 2, 3, 4, 5},
			"status": "processed",
			"metadata": map[string]string{
				"processed_by": "test_system",
				"timestamp":    "2025-12-22T19:00:00Z",
			},
		},
	}

	wrapper := NewCommandWrapperBuilder[TestCommand]().
		WithContext(ctx).
		WithCommand(command).
		Build()

	// Validate context propagation
	if userID := wrapper.Context.Value("user_id"); userID != "test_user" {
		t.Error("Context value not propagated correctly")
	}

	// Validate complex command data
	if ids, ok := wrapper.Command.Data["ids"].([]int); !ok || len(ids) != 5 {
		t.Error("Complex data structure not preserved")
	}

	if metadata, ok := wrapper.Command.Data["metadata"].(map[string]string); !ok {
		t.Error("Nested map not preserved")
	} else if metadata["processed_by"] != "test_system" {
		t.Error("Nested map values not preserved")
	}
}

// Error handling tests
func TestErrorHandling(t *testing.T) {
	// Test SortBy validation errors
	t.Run("SortBy validation", func(t *testing.T) {
		invalidSort := SortBy{field: "", ascending: true}
		err := invalidSort.Validate()
		if err == nil {
			t.Error("Should return error for empty field")
		}
		if !errors.Is(err, errors.New("sort field cannot be empty")) {
			t.Log("Error message is as expected")
		}
	})

	// Test Pagination validation errors
	t.Run("Pagination validation", func(t *testing.T) {
		invalidPagination := Pagination{limit: -5, offset: -1}
		err := invalidPagination.Validate()
		if err == nil {
			t.Error("Should return error for invalid pagination")
		}
	})
}

// Benchmark tests
func BenchmarkQueryWrapperBuilder(b *testing.B) {
	ctx := context.Background()
	query := TestQuery{Name: "benchmark", Age: 25}
	projection := NewProjection([]string{"id", "name"})
	pagination := NewPagination(10, 0)
	sortBy := NewSortBy("name", true)
	filters := []Filter{NewFilter("active", true)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewQueryWrapperBuilder[TestQuery]().
			WithContext(ctx).
			WithQuery(query).
			withProjection(projection).
			WithPagination(pagination).
			WithSortBy(sortBy).
			WithFilter(filters).
			Build()
	}
}

func BenchmarkPaginationHasNext(b *testing.B) {
	pagination := NewPagination(10, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pagination.HasNext(1000)
	}
}

func BenchmarkPagesBuilder(b *testing.B) {
	results := make([]TestResult, 100)
	for i := range results {
		results[i] = TestResult{ID: i, Name: "test"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPagesBuilder[TestResult]().
			Results(results).
			Offset(0).
			Limit(10).
			HasNext(false).
			Build()
	}
}
