package store

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

// Store elements of type T in a database.
type Store[T Storable] interface {
	Insert(context.Context, T) error
	Get(context.Context, spanner.Key) (T, error)
	List(context.Context, int64, int64) ([]T, error)
}

// Storable in a database.
type Storable interface {
	Key() spanner.Key
	Table() string
	List(int64, int64) spanner.Statement
}

// getColumnsForType helps to simplify the Spanner logic for what columns to retrieve.
func getColumnsForType(x any) []string {
	var columns []string

	t := reflect.TypeOf(x)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		columns = append(columns, field.Name)
	}

	return columns
}

// Spanner store for elements.
type Spanner[T Storable] struct {
	client *spanner.Client
}

// InSpanner stores items backed by a Spanner database.
func InSpanner[T Storable](client *spanner.Client) *Spanner[T] {
	return &Spanner[T]{client: client}
}

// Insert an item into the store.
func (s *Spanner[T]) Insert(ctx context.Context, item T) error {
	var mutations []*spanner.Mutation

	m, err := spanner.InsertOrUpdateStruct(item.Table(), item)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	mutations = append(mutations, m)

	if _, err := s.client.Apply(ctx, mutations); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}

// Get an element from the store by key.
func (s *Spanner[T]) Get(ctx context.Context, key spanner.Key) (T, error) {
	var item T

	row, err := s.client.Single().ReadRow(ctx, item.Table(), key, getColumnsForType(item))
	if err != nil {
		return item, fmt.Errorf("get: %w", err)
	}

	if err := row.ToStruct(&item); err != nil {
		return item, fmt.Errorf("get: %w", err)
	}

	return item, nil
}

// List elements with limit and offset for pagination.
func (s *Spanner[T]) List(ctx context.Context, limit, offset int64) ([]T, error) {
	var item T

	var items []T

	iter := s.client.Single().Query(ctx, item.List(limit, offset))

	for {
		row, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("list: %w", err)
		}

		if err := row.ToStruct(&item); err != nil {
			return nil, fmt.Errorf("list: %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}
