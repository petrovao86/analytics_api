package config

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrWrongType = errors.New("wrong type")
)

type IConfigurable interface {
	Configure(IReader) error
}

type IReader interface {
	Get(key string) (any, bool)
	Sub(key string) (IReader, bool)
	Map() map[string]any
}

func WithReader[T any](cr IReader) func(T) error {
	return func(o T) error {
		if i, ok := any(o).(IConfigurable); ok {
			return i.Configure(cr)
		}
		return nil
	}
}

func Get[T any](c IReader, key string) (T, error) {
	var v T
	rawV, ok := c.Get(key)
	if !ok {
		return v, fmt.Errorf("config key \"%v\"; %w ", key, ErrNotFound)
	}

	v, ok = rawV.(T)
	if !ok {
		return v, fmt.Errorf("config key \"%v\"; %w: %v %v", key, ErrWrongType, rawV, reflect.TypeOf(rawV).String())
	}
	return v, nil
}

func Sub(c IReader, key string) (IReader, error) {
	v, ok := c.Sub(key)
	if !ok {
		return nil, fmt.Errorf("config key \"%v\"; %w", key, ErrNotFound)
	}
	return v, nil
}
