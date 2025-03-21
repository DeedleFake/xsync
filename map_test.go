package xsync_test

import (
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"deedles.dev/xsync"
)

type mapOp string

const (
	opLoad             = mapOp("Load")
	opStore            = mapOp("Store")
	opLoadOrStore      = mapOp("LoadOrStore")
	opLoadAndDelete    = mapOp("LoadAndDelete")
	opDelete           = mapOp("Delete")
	opSwap             = mapOp("Swap")
	opCompareAndSwap   = mapOp("CompareAndSwap")
	opCompareAndDelete = mapOp("CompareAndDelete")
	opClear            = mapOp("Clear")
)

var mapOps = [...]mapOp{
	opLoad,
	opStore,
	opLoadOrStore,
	opLoadAndDelete,
	opDelete,
	opSwap,
	opCompareAndSwap,
	opCompareAndDelete,
	opClear,
}

type mapCall struct {
	op   mapOp
	k, v any
}

func (c mapCall) apply(m *xsync.Map[any, any]) (any, bool) {
	switch c.op {
	case opLoad:
		return m.Load(c.k)
	case opStore:
		m.Store(c.k, c.v)
		return nil, false
	case opLoadOrStore:
		return m.LoadOrStore(c.k, c.v)
	case opLoadAndDelete:
		return m.LoadAndDelete(c.k)
	case opDelete:
		m.Delete(c.k)
		return nil, false
	case opSwap:
		return m.Swap(c.k, c.v)
	case opCompareAndSwap:
		if m.CompareAndSwap(c.k, c.v, rand.Int()) {
			m.Delete(c.k)
			return c.v, true
		}
		return nil, false
	case opCompareAndDelete:
		if m.CompareAndDelete(c.k, c.v) {
			if _, ok := m.Load(c.k); !ok {
				return nil, true
			}
		}
		return nil, false
	case opClear:
		m.Clear()
		return nil, false
	default:
		panic("invalid mapOp")
	}
}

type mapResult struct {
	value any
	ok    bool
}

func randValue(r *rand.Rand) any {
	b := make([]byte, r.Intn(4))
	for i := range b {
		b[i] = 'a' + byte(rand.Intn(26))
	}
	return string(b)
}

func (mapCall) Generate(r *rand.Rand, size int) reflect.Value {
	c := mapCall{op: mapOps[rand.Intn(len(mapOps))], k: randValue(r)}
	switch c.op {
	case opStore, opLoadOrStore:
		c.v = randValue(r)
	}
	return reflect.ValueOf(c)
}

func applyCalls(m *xsync.Map[any, any], calls []mapCall) (results []mapResult, final map[any]any) {
	for _, c := range calls {
		v, ok := c.apply(m)
		results = append(results, mapResult{v, ok})
	}

	final = make(map[any]any)
	m.Range(func(k, v any) bool {
		final[k] = v
		return true
	})

	return results, final
}

func applyMap(calls []mapCall) ([]mapResult, map[any]any) {
	return applyCalls(new(xsync.Map[any, any]), calls)
}

func TestConcurrentRange(t *testing.T) {
	const mapSize = 1 << 10

	m := new(sync.Map)
	for n := int64(1); n <= mapSize; n++ {
		m.Store(n, int64(n))
	}

	done := make(chan struct{})
	var wg sync.WaitGroup
	defer func() {
		close(done)
		wg.Wait()
	}()
	for g := int64(runtime.GOMAXPROCS(0)); g > 0; g-- {
		r := rand.New(rand.NewSource(g))
		wg.Add(1)
		go func(g int64) {
			defer wg.Done()
			for i := int64(0); ; i++ {
				select {
				case <-done:
					return
				default:
				}
				for n := int64(1); n < mapSize; n++ {
					if r.Int63n(mapSize) == 0 {
						m.Store(n, n*i*g)
					} else {
						m.Load(n)
					}
				}
			}
		}(g)
	}

	iters := 1 << 10
	if testing.Short() {
		iters = 16
	}
	for n := iters; n > 0; n-- {
		seen := make(map[int64]bool, mapSize)

		m.Range(func(ki, vi any) bool {
			k, v := ki.(int64), vi.(int64)
			if v%k != 0 {
				t.Fatalf("while Storing multiples of %v, Range saw value %v", k, v)
			}
			if seen[k] {
				t.Fatalf("Range visited key %v twice", k)
			}
			seen[k] = true
			return true
		})

		if len(seen) != mapSize {
			t.Fatalf("Range visited %v elements of %v-element Map", len(seen), mapSize)
		}
	}
}

func TestIssue40999(t *testing.T) {
	var m sync.Map

	m.Store(nil, struct{}{})

	var finalized uint32

	for atomic.LoadUint32(&finalized) == 0 {
		p := new(int)
		runtime.SetFinalizer(p, func(*int) {
			atomic.AddUint32(&finalized, 1)
		})
		m.Store(p, struct{}{})
		m.Delete(p)
		runtime.GC()
	}
}

func TestMapRangeNestedCall(t *testing.T) { // Issue 46399
	var m sync.Map
	for i, v := range [3]string{"hello", "world", "Go"} {
		m.Store(i, v)
	}
	m.Range(func(key, value any) bool {
		m.Range(func(key, value any) bool {
			if v, ok := m.Load(key); !ok || !reflect.DeepEqual(v, value) {
				t.Fatalf("Nested Range loads unexpected value, got %+v want %+v", v, value)
			}

			if _, loaded := m.LoadOrStore(42, "dummy"); loaded {
				t.Fatalf("Nested Range loads unexpected value, want store a new value")
			}

			val := "sync.Map"
			m.Store(42, val)
			if v, loaded := m.LoadAndDelete(42); !loaded || !reflect.DeepEqual(v, val) {
				t.Fatalf("Nested Range loads unexpected value, got %v, want %v", v, val)
			}
			return true
		})

		m.Delete(key)
		return true
	})

	length := 0
	m.Range(func(key, value any) bool {
		length++
		return true
	})

	if length != 0 {
		t.Fatalf("Unexpected sync.Map size, got %v want %v", length, 0)
	}
}

func TestCompareAndSwap_NonExistingKey(t *testing.T) {
	m := &sync.Map{}
	if m.CompareAndSwap(m, nil, 42) {
		t.Fatalf("CompareAndSwap on a non-existing key succeeded")
	}
}

func TestMapRangeNoAllocations(t *testing.T) { // Issue 62404
	var m sync.Map
	allocs := testing.AllocsPerRun(10, func() {
		m.Range(func(key, value any) bool {
			return true
		})
	})
	if allocs > 0 {
		t.Errorf("AllocsPerRun of m.Range = %v; want 0", allocs)
	}
}

func TestConcurrentClear(t *testing.T) {
	var m sync.Map

	wg := sync.WaitGroup{}
	wg.Add(30) // 10 goroutines for writing, 10 goroutines for reading, 10 goroutines for waiting

	for i := range 10 {
		go func(k, v int) {
			defer wg.Done()
			m.Store(k, v)
		}(i, i*10)
	}

	for i := range 10 {
		go func(k int) {
			defer wg.Done()
			if value, ok := m.Load(k); ok {
				t.Logf("Key: %v, Value: %v\n", k, value)
			} else {
				t.Logf("Key: %v not found\n", k)
			}
		}(i)
	}

	for range 10 {
		go func() {
			defer wg.Done()
			m.Clear()
		}()
	}

	wg.Wait()

	m.Clear()

	m.Range(func(k, v any) bool {
		t.Errorf("after Clear, Map contains (%v, %v); expected to be empty", k, v)

		return true
	})
}

func TestMapClearNoAllocations(t *testing.T) {
	t.SkipNow()

	var m sync.Map
	allocs := testing.AllocsPerRun(10, func() {
		m.Clear()
	})
	if allocs > 0 {
		t.Errorf("AllocsPerRun of m.Clear = %v; want 0", allocs)
	}
}
