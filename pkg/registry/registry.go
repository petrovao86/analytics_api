package registry

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"sync"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type IRegistry[T any] interface {
	Register(string, T) (T, error)
	Deregister(string) error
	Get(string) (T, error)
	ForEach(f func(key string, value T) bool)
	Update(key string, f func(value T) (T, error)) (T, error)
	All() map[string]T
}

type registry[T any] struct {
	data map[string]T
	m    sync.RWMutex
}

func NewRegistry[T any]() IRegistry[T] {
	return &registry[T]{data: make(map[string]T)}
}

func (r *registry[T]) Get(name string) (T, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	if e, found := r.data[name]; found {
		return e, nil
	}
	var result T
	return result, fmt.Errorf("key \"%v\"; %w ", name, ErrNotFound)
}

func (r *registry[T]) Register(name string, o T) (T, error) {
	r.m.Lock()
	defer r.m.Unlock()
	result, exists := r.data[name]
	if exists {
		return result, ErrAlreadyExists
	}

	r.data[name] = o
	return o, nil
}

func (r *registry[T]) All() map[string]T {
	r.m.RLock()
	defer r.m.RUnlock()
	return maps.Clone(r.data)
}

func (r *registry[T]) ForEach(f func(key string, value T) bool) {
	r.m.RLock()
	defer r.m.RUnlock()
	for k, v := range r.data {
		if next := f(k, v); !next {
			return
		}
	}
}

func (r *registry[T]) Update(key string, f func(value T) (T, error)) (T, error) {
	r.m.Lock()
	defer r.m.Unlock()
	newValue, err := f(r.data[key])
	if err != nil {
		return newValue, err
	}
	if isNil(newValue) {
		delete(r.data, key)
		return newValue, nil
	}

	r.data[key] = newValue
	return newValue, nil
}

func (r *registry[T]) Deregister(name string) error {
	r.m.Lock()
	defer r.m.Unlock()
	if _, exists := r.data[name]; !exists {
		return ErrNotFound
	}

	delete(r.data, name)
	return nil
}

func isNil[T any](t T) bool {
	v := reflect.ValueOf(t)
	kind := v.Kind()
	return (kind == reflect.Ptr ||
		kind == reflect.Interface ||
		kind == reflect.Slice ||
		kind == reflect.Map ||
		kind == reflect.Chan ||
		kind == reflect.Func) &&
		v.IsNil()
}
