package xmap

// DirtyMarker abstracts a type that can mark a specific dirty bit.
// Any component embedding a tracker that exposes MarkDirty(int64) can implement this.
type DirtyMarker interface {
	MarkDirty(bit int64)
}

// MapAccessor abstracts the access to a specific map field and its dirty marker.
type MapAccessor[K comparable, V any] struct {
	zero          V
	m             *map[K]V
	changeTracker *FinegrainedChangeTracker[K]
	// Backward-compatible ad-hoc dirty callback
	markDirty func()
	// Preferred: interface-based marker to avoid closure allocations
	marker   DirtyMarker
	dirtyBit int64
}

// NewMapAccessor creates an accessor bound to a component field map pointer.
// All write operations will modify the underlying field map and invoke markDirty (if provided).
func NewMapAccessor[K comparable, V any](m *map[K]V, markDirty func()) MapAccessor[K, V] {
	return MapAccessor[K, V]{m: m, markDirty: markDirty}
}

// NewMapAccessorWithMarker creates an accessor bound to a component field map pointer
// and a DirtyMarker with a concrete dirty bit. This avoids closure allocations.
func NewMapAccessorWithMarker[K comparable, V any](m *map[K]V, marker DirtyMarker, dirtyBit int64) MapAccessor[K, V] {
	return MapAccessor[K, V]{m: m, marker: marker, dirtyBit: dirtyBit, changeTracker: NewFinegrainedChangeTracker[K]()}
}

func (op MapAccessor[K, V]) Get(key K) (V, bool) {
	if op.m == nil || *op.m == nil {
		return op.zero, false
	}
	v, ok := (*op.m)[key]
	return v, ok
}

func (op MapAccessor[K, V]) Has(key K) bool {
	if op.m == nil || *op.m == nil {
		return false
	}
	_, ok := (*op.m)[key]
	return ok
}

func (op MapAccessor[K, V]) ensureMapNoDirty() {
	if op.m == nil {
		return
	}
	if *op.m == nil {
		*op.m = make(map[K]V)
	}
}

func (op MapAccessor[K, V]) Set(key K, value V) {
	op.ensureMapNoDirty()
	if op.m != nil {
		(*op.m)[key] = value
		op.changeTracker.TrackSet(key)
		op.doMarkDirty()
	}
}

// Upsert sets the value and returns (previous, existed).
func (op MapAccessor[K, V]) Upsert(key K, value V) (V, bool) {
	op.ensureMapNoDirty()
	if op.m == nil {
		return op.zero, false
	}
	prev, existed := (*op.m)[key]
	(*op.m)[key] = value
	op.changeTracker.TrackSet(key)
	op.doMarkDirty()
	return prev, existed
}

func (op MapAccessor[K, V]) Delete(key K) {
	if op.m == nil || *op.m == nil {
		return
	}
	if _, ok := (*op.m)[key]; ok {
		delete(*op.m, key)
		op.changeTracker.TrackUnset(key)
		op.doMarkDirty()
	}
}

func (op MapAccessor[K, V]) Clear() {
	if op.m == nil || *op.m == nil {
		return
	}
	for k := range *op.m {
		delete(*op.m, k)
		op.changeTracker.TrackUnset(k)
	}

	op.doMarkDirty()
}

func (op MapAccessor[K, V]) Range(f func(key K, value V) bool) {
	if op.m == nil || *op.m == nil || f == nil {
		return
	}
	for key, value := range *op.m {
		if !f(key, value) {
			break
		}
	}
}

func (op MapAccessor[K, V]) Len() int {
	if op.m == nil || *op.m == nil {
		return 0
	}
	return len(*op.m)
}

// Ensure returns the underlying map, allocating it if necessary.
// Does not mark dirty by itself.
func (op MapAccessor[K, V]) Ensure() map[K]V {
	if op.m == nil {
		return nil
	}
	if *op.m == nil {
		*op.m = make(map[K]V)
	}
	return *op.m
}

func (op MapAccessor[K, V]) SetAll(entries map[K]V) {
	if entries == nil {
		return
	}
	op.ensureMapNoDirty()
	if op.m == nil {
		return
	}
	for k, v := range entries {
		(*op.m)[k] = v
		op.changeTracker.TrackSet(k)
	}
	op.doMarkDirty()
}

func (op MapAccessor[K, V]) DeleteAll(keys ...K) int {
	if op.m == nil || *op.m == nil || len(keys) == 0 {
		return 0
	}
	count := 0
	for _, k := range keys {
		if _, ok := (*op.m)[k]; ok {
			delete(*op.m, k)
			op.changeTracker.TrackUnset(k)
			count++
		}
	}
	if count > 0 {
		op.doMarkDirty()
	}
	return count
}

func (op MapAccessor[K, V]) DeepCopy() map[K]V {
	if op.m == nil || *op.m == nil {
		return nil
	}
	dup := make(map[K]V, len(*op.m))
	for k, v := range *op.m {
		dup[k] = v
	}
	return dup
}

func (op MapAccessor[K, V]) Pop(k K) (V, bool) {
	if op.m == nil || *op.m == nil {
		return op.zero, false
	}
	v, ok := (*op.m)[k]
	if ok {
		delete(*op.m, k)
		op.changeTracker.TrackUnset(k)
		op.doMarkDirty()
	}
	return v, ok
}

func (op MapAccessor[K, V]) Keys() []K {
	if op.m == nil || *op.m == nil {
		return nil
	}
	keys := make([]K, 0, len(*op.m))
	for k := range *op.m {
		keys = append(keys, k)
	}
	return keys
}

func (op MapAccessor[K, V]) Values() []V {
	if op.m == nil || *op.m == nil {
		return nil
	}
	vals := make([]V, 0, len(*op.m))
	for _, v := range *op.m {
		vals = append(vals, v)
	}
	return vals
}

// doMarkDirty triggers the appropriate dirty mechanism, preferring the interface-based marker.
func (op MapAccessor[K, V]) doMarkDirty() {
	if op.marker != nil {
		op.marker.MarkDirty(op.dirtyBit)
		return
	}
	if op.markDirty != nil {
		op.markDirty()
	}
}
