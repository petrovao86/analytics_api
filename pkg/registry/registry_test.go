package registry

import (
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry[int]()
	val, err := r.Register("test", 1)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, val)
	}
	val, err = r.Register("test", 2)
	if assert.ErrorIs(t, err, ErrAlreadyExists) {
		assert.Equal(t, 1, val)
	}

	val, err = r.Register("test2", 2)
	if assert.NoError(t, err) {
		assert.Equal(t, 2, val)
	}

}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry[int]()
	_, err := r.Get("unknown")
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = r.Register("test", 1)
	if err == nil {
		_, err := r.Get("unknown")
		assert.ErrorIs(t, err, ErrNotFound)

		val, err := r.Get("test")
		if assert.NoError(t, err) {
			assert.Equal(t, 1, val)
		}
	}
}

func TestRegistry_Deregister(t *testing.T) {
	r := NewRegistry[int]()
	assert.ErrorIs(t, r.Deregister("unknown"), ErrNotFound)
	_, err := r.Register("test", 1)
	if err == nil {
		assert.ErrorIs(t, r.Deregister("unknown"), ErrNotFound)
		assert.NoError(t, r.Deregister("test"))
		_, err := r.Get("test")
		assert.ErrorIs(t, err, ErrNotFound)
	}
}

func TestRegistry_All(t *testing.T) {
	r := NewRegistry[int]()
	assert.Len(t, r.All(), 0)
	_, err := r.Register("test", 1)
	if err == nil {
		assert.Equal(t, map[string]int{"test": 1}, r.All())
		err = r.Deregister("test")
		if err == nil {
			assert.Len(t, r.All(), 0)
		}
	}
}

func TestRegistry_ForEach(t *testing.T) {
	r := NewRegistry[int]()
	visits := 0
	r.ForEach(func(_ string, _ int) bool {
		visits++
		return true
	})
	assert.Equal(t, 0, visits)

	visits = 0
	_, err := r.Register("test", 1)
	if err == nil {
		r.ForEach(func(_ string, _ int) bool {
			visits++
			return true
		})
		assert.Equal(t, 1, visits)
	}
	visits = 0
	_, err = r.Register("test2", 2)
	if err == nil {
		r.ForEach(func(_ string, _ int) bool {
			visits++
			return true
		})
		assert.Equal(t, 2, visits)

		visits = 0
		r.ForEach(func(_ string, _ int) bool {
			visits++
			return false
		})
		assert.Equal(t, 1, visits)

	}
}

func TestRegistry_Update(t *testing.T) {
	type testCase struct {
		initVal           map[string][]bool
		key               string
		updateF           func(value []bool) ([]bool, error)
		expectedUpdateErr error
		expectedErr       error
		expectedVal       []bool
	}
	testErr := errors.New("")
	testCases := map[string]testCase{
		"create_key": {
			initVal: make(map[string][]bool), key: "new_key",
			updateF:           func(value []bool) ([]bool, error) { return []bool{true}, nil },
			expectedUpdateErr: nil, expectedErr: nil, expectedVal: []bool{true},
		},
		"create_err": {
			initVal: make(map[string][]bool), key: "new_key_err",
			updateF:           func(value []bool) ([]bool, error) { return nil, testErr },
			expectedUpdateErr: testErr, expectedErr: ErrNotFound, expectedVal: nil,
		},
		"delete_unknown_key": {
			initVal: make(map[string][]bool), key: "delete_key",
			updateF:           func(value []bool) ([]bool, error) { return nil, nil },
			expectedUpdateErr: nil, expectedErr: ErrNotFound, expectedVal: nil,
		},
		"delete_existing_key": {
			initVal: map[string][]bool{"delete_key": {true}}, key: "delete_key",
			updateF:           func(value []bool) ([]bool, error) { return nil, nil },
			expectedUpdateErr: nil, expectedErr: ErrNotFound, expectedVal: nil,
		},
		"delete_err": {
			initVal: map[string][]bool{"delete_key_err": {true}}, key: "delete_key_err",
			updateF:           func(value []bool) ([]bool, error) { return nil, testErr },
			expectedUpdateErr: testErr, expectedErr: nil, expectedVal: []bool{true},
		},
		"update_key": {
			initVal: map[string][]bool{"update_key": {true}}, key: "update_key",
			updateF:           func(value []bool) ([]bool, error) { return []bool{false}, nil },
			expectedUpdateErr: nil, expectedErr: nil, expectedVal: []bool{false},
		},
		"update_error": {
			initVal: map[string][]bool{"update_key": {true}}, key: "update_key",
			updateF:           func(value []bool) ([]bool, error) { return nil, testErr },
			expectedUpdateErr: testErr, expectedErr: nil, expectedVal: []bool{true},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(test *testing.T) {
			reg := NewRegistry[[]bool]()
			for key, val := range tc.initVal {
				reg.Register(key, val)
			}
			_, err := reg.Update(tc.key, tc.updateF)
			assert.ErrorIs(test, err, tc.expectedUpdateErr)
			val, err := reg.Get(tc.key)
			assert.ErrorIs(test, err, tc.expectedErr)
			assert.Equal(test, tc.expectedVal, val)

		})
	}
}

func TestRegistry_UpdateParallel(t *testing.T) {
	r := NewRegistry[[]string]()
	entryCnt := 50
	keyCnt := 100
	doublesCnt := 10
	errCh := make(chan error, entryCnt*keyCnt)
	errS := make([]error, 0)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for err := range errCh {
			if err != nil {
				errS = append(errS, err)
				if len(errS) >= doublesCnt*entryCnt {
					break
				}
			}
		}
		wg.Done()
	}()
	for entryId := 0; entryId < entryCnt; entryId++ {
		wg.Add(1)
		go func(reg IRegistry[[]string], eId int) {
			for key := 0; key < keyCnt; key++ {
				if _, err := reg.Update(
					strconv.Itoa(key),
					func(entries []string) ([]string, error) {
						if entries == nil {
							entries = make([]string, 0)
						}
						entry := strconv.Itoa(eId)
						pos := slices.Index(entries, entry)
						if pos >= 0 {
							return nil, ErrAlreadyExists
						}
						entries = append(entries, entry)
						return entries, nil
					},
				); err != nil {
					errCh <- err
				}
			}
			wg.Done()
		}(r, entryId)

		wg.Add(1)
		go func(reg IRegistry[[]string], eId int) {
			for key := 0; key < doublesCnt; key++ {
				if _, err := reg.Update(
					strconv.Itoa(key),
					func(entries []string) ([]string, error) {
						if entries == nil {
							entries = make([]string, 0)
						}
						entry := strconv.Itoa(eId)
						pos := slices.Index(entries, entry)
						if pos >= 0 {
							return nil, ErrAlreadyExists
						}
						entries = append(entries, entry)
						return entries, nil
					},
				); err != nil {
					errCh <- err
				}
			}
			wg.Done()
		}(r, entryId)
	}
	wg.Wait()
	close(errCh)

	if assert.Len(t, errS, doublesCnt*entryCnt) {
		endpoints := r.All()
		if assert.Len(t, endpoints, keyCnt) {
			for e := 0; e < keyCnt; e++ {
				nodes := endpoints[strconv.Itoa(e)]
				assert.Len(t, nodes, entryCnt)
			}
		}
	}
}
