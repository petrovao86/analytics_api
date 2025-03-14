package config

type mapReader map[string]any

var _ IReader = (*mapReader)(nil)

func NewMapReader(m map[string]any) IReader {
	return mapReader(m)
}

func (cr mapReader) Get(field string) (any, bool) {
	v, ok := cr[field]
	return v, ok
}

func (cr mapReader) Sub(field string) (IReader, bool) {
	s, ok := cr[field]
	if !ok {
		return nil, false
	}
	sub, ok := s.(map[string]any)
	if !ok {
		return nil, false
	}
	return NewMapReader(sub), true
}

func (cr mapReader) Map() map[string]any {
	return cr
}
