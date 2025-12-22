package wrapper

import (
	"context"
	"errors"
)

type Void struct{}
type Empty struct{}

type QueryWrapper[Q any] struct {
	Context    context.Context
	Query      Q
	projection Projection
	pagination Pagination
	sortBy     SortBy
	filter     []Filter
}

// Projection returns the projection
func (qw QueryWrapper[Q]) Projection() Projection {
	return qw.projection
}

// Pagination returns the pagination
func (qw QueryWrapper[Q]) Pagination() Pagination {
	return qw.pagination
}

// SortBy returns the sort configuration
func (qw QueryWrapper[Q]) SortBy() SortBy {
	return qw.sortBy
}

// Filter returns the filters
func (qw QueryWrapper[Q]) Filter() []Filter {
	return qw.filter
}

func NewQueryWrapper[T any](ctx context.Context, query T, projection Projection, pagination Pagination, sortBy SortBy, filter []Filter) QueryWrapper[T] {
	return QueryWrapper[T]{
		Context:    ctx,
		Query:      query,
		projection: projection,
		pagination: pagination,
		sortBy:     sortBy,
		filter:     filter,
	}
}

type QueryWrapperBuilder[Q any] struct {
	ctx        context.Context
	query      Q
	projection Projection
	pagination Pagination
	sortBy     SortBy
	filter     []Filter
}

func NewQueryWrapperBuilder[T any]() *QueryWrapperBuilder[T] {
	return &QueryWrapperBuilder[T]{}
}

func (b *QueryWrapperBuilder[T]) WithContext(ctx context.Context) *QueryWrapperBuilder[T] {
	b.ctx = ctx
	return b
}

func (b *QueryWrapperBuilder[T]) WithQuery(query T) *QueryWrapperBuilder[T] {
	b.query = query
	return b
}

func (b *QueryWrapperBuilder[T]) withProjection(projection Projection) *QueryWrapperBuilder[T] {
	b.projection = projection
	return b
}

func (b *QueryWrapperBuilder[T]) WithPagination(pagination Pagination) *QueryWrapperBuilder[T] {
	b.pagination = pagination
	return b
}

func (b *QueryWrapperBuilder[T]) WithSortBy(sortBy SortBy) *QueryWrapperBuilder[T] {
	b.sortBy = sortBy
	return b
}

func (b *QueryWrapperBuilder[T]) WithFilter(filter []Filter) *QueryWrapperBuilder[T] {
	b.filter = filter
	return b
}

func (b *QueryWrapperBuilder[T]) Build() QueryWrapper[T] {
	return QueryWrapper[T]{
		Context:    b.ctx,
		Query:      b.query,
		projection: b.projection,
		pagination: b.pagination,
		sortBy:     b.sortBy,
		filter:     b.filter,
	}
}

type CommandWrapper[C any] struct {
	Context context.Context
	Command C
}

func NewCommandWrapper[T any](ctx context.Context, command T) CommandWrapper[T] {
	return CommandWrapper[T]{
		Context: ctx,
		Command: command,
	}
}

type CommandWrapperBuilder[C any] struct {
	ctx     context.Context
	command C
}

func NewCommandWrapperBuilder[C any]() *CommandWrapperBuilder[C] {
	return &CommandWrapperBuilder[C]{}
}

func (b *CommandWrapperBuilder[C]) WithContext(ctx context.Context) *CommandWrapperBuilder[C] {
	b.ctx = ctx
	return b
}

func (b *CommandWrapperBuilder[C]) WithCommand(command C) *CommandWrapperBuilder[C] {
	b.command = command
	return b
}

func (b *CommandWrapperBuilder[C]) Build() CommandWrapper[C] {
	return CommandWrapper[C]{
		Context: b.ctx,
		Command: b.command,
	}
}

type SortBy struct {
	field     string
	ascending bool
}

func NewSortBy(field string, ascending bool) SortBy {
	return SortBy{
		field:     field,
		ascending: ascending,
	}
}

func NewAscendingSortBy(field string) SortBy {
	return SortBy{
		field:     field,
		ascending: true,
	}
}

func NewDescendingSortBy(field string) SortBy {
	return SortBy{
		field:     field,
		ascending: false,
	}
}

func (s SortBy) Field() string {
	return s.field
}

func (s SortBy) IsAscending() bool {
	return s.ascending
}

func (s SortBy) Validate() error {
	if s.field == "" {
		return errors.New("sort field cannot be empty")
	}
	return nil
}

type Pagination struct {
	limit  int
	offset int
}

func NewPagination(limit, offset int) Pagination {
	return Pagination{
		limit:  limit,
		offset: offset,
	}
}

func NewFirstPagePagination() Pagination {
	return Pagination{
		limit:  10,
		offset: 0,
	}
}

func (p Pagination) Limit() int {
	return p.limit
}

func (p Pagination) Offset() int {
	return p.offset
}

func (p Pagination) HasNext(totalCount int) bool {
	return p.offset+p.limit < totalCount
}

func (p Pagination) NextPage() Pagination {
	return Pagination{
		limit:  p.limit,
		offset: p.offset + p.limit,
	}
}

func (p Pagination) Validate() error {
	if p.limit <= 0 {
		return errors.New("limit must be positive")
	}
	if p.offset < 0 {
		return errors.New("offset cannot be negative")
	}
	return nil
}

type Filter struct {
	field string
	value any
}

func NewFilter(field string, value any) Filter {
	return Filter{
		field: field,
		value: value,
	}
}

func (f Filter) Field() string {
	return f.field
}

func (f Filter) Value() any {
	return f.value
}

type Projection struct {
	fields []string
}

// NewProjection creates a new Projection with the given fields
func NewProjection(fields []string) Projection {
	return Projection{
		fields: fields,
	}
}

// NewEmptyProjection creates an empty projection
func NewEmptyProjection() Projection {
	return Projection{
		fields: []string{},
	}
}

func (p Projection) Fields() []string {
	return p.fields
}

type Page[R any] struct {
	Results []R
	Offset  int
	Limit   int
	HasNext bool
}

func (r *Page[R]) Next() bool {
	return r.HasNext
}

func (r *Page[R]) First() bool {
	return r.Offset == 0
}

func (r *Page[R]) HasData() bool {
	return len(r.Results) > 0
}

type PagesBuilder[R any] struct {
	results []R
	offset  int
	limit   int
	hasNext bool
}

func NewPagesBuilder[R any]() *PagesBuilder[R] {
	return &PagesBuilder[R]{}
}

func (p *PagesBuilder[R]) Results(results []R) *PagesBuilder[R] {
	p.results = results
	return p
}

func (p *PagesBuilder[R]) Offset(offset int) *PagesBuilder[R] {
	p.offset = offset
	return p
}

func (p *PagesBuilder[R]) Limit(limit int) *PagesBuilder[R] {
	p.limit = limit
	return p
}

func (p *PagesBuilder[R]) HasNext(hasNext bool) *PagesBuilder[R] {
	p.hasNext = hasNext
	return p
}

func (p *PagesBuilder[R]) Build() Page[R] {
	return Page[R]{
		Results: p.results,
		Offset:  p.offset,
		Limit:   p.limit,
		HasNext: p.hasNext,
	}
}
