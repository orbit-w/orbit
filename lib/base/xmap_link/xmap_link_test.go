package xmaplink

import (
	"testing"

	dt "gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
)

// ============================================
// Mock 对象用于测试
// ============================================

// MockPbData 模拟 protobuf 数据
type MockPbData struct {
	Value int32
}

// MockWrapper 模拟包装对象
type MockWrapper struct {
	pb               *MockPbData
	tracker          dt.DirtyTracker
	factsAccessor    xmap.FactsAccessor[int64]
	factsAccessorKey int64
}

func NewMockWrapper(pb *MockPbData) *MockWrapper {
	return &MockWrapper{
		pb: pb,
	}
}

// 实现 Linkable 接口
func (w *MockWrapper) Link(parent *dt.DirtyTracker, parentBit int64) {
	w.tracker.Link(parent, parentBit)
}

func (w *MockWrapper) LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64, factsAccessor xmap.FactsAccessor[int64], key int64) {
	w.factsAccessor = factsAccessor
	w.factsAccessorKey = key
	w.tracker.Link(parent, parentBit)
}

func (w *MockWrapper) Unlink() {
	w.tracker.Unlink()
	w.factsAccessor = nil
	w.factsAccessorKey = 0
}

func (w *MockWrapper) GetDirtyTracker() *dt.DirtyTracker {
	return &w.tracker
}

func (w *MockWrapper) MarkDirty(dirtyBit int64) {
	w.tracker.MarkDirty(dirtyBit)
	if w.factsAccessor != nil {
		w.factsAccessor.TrackSet(w.factsAccessorKey)
	}
}

func (w *MockWrapper) IsLinked() bool {
	return w.factsAccessor != nil
}

// ============================================
// 测试用例
// ============================================

func TestXMapLink_New(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}
	pbMap[2] = &MockPbData{Value: 200}

	tracker := &dt.DirtyTracker{}
	const dirtyBit = int64(1 << 0)

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		dirtyBit,
		NewMockWrapper,
		true,
	)

	// 验证初始化
	if link == nil {
		t.Fatal("NewXMapLinkWithParent returned nil")
	}

	// 验证长度
	if link.Len() != 2 {
		t.Errorf("Expected length 2, got %d", link.Len())
	}

	// 验证已有对象被正确包装和链接
	wrapper1, ok := link.Get(1)
	if !ok {
		t.Error("Expected key 1 to exist")
	}
	if wrapper1.pb.Value != 100 {
		t.Errorf("Expected value 100, got %d", wrapper1.pb.Value)
	}
	if !wrapper1.IsLinked() {
		t.Error("Expected wrapper1 to be linked")
	}

	wrapper2, ok := link.Get(2)
	if !ok {
		t.Error("Expected key 2 to exist")
	}
	if wrapper2.pb.Value != 200 {
		t.Errorf("Expected value 200, got %d", wrapper2.pb.Value)
	}
	if !wrapper2.IsLinked() {
		t.Error("Expected wrapper2 to be linked")
	}
}

func TestXMapLink_Get(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}

	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 测试获取存在的key
	wrapper, ok := link.Get(1)
	if !ok {
		t.Error("Expected key 1 to exist")
	}
	if wrapper.pb.Value != 100 {
		t.Errorf("Expected value 100, got %d", wrapper.pb.Value)
	}

	// 测试获取不存在的key
	_, ok = link.Get(999)
	if ok {
		t.Error("Expected key 999 to not exist")
	}
}

func TestXMapLink_Set(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}
	const dirtyBit = int64(1 << 0)

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		dirtyBit,
		NewMockWrapper,
		true,
	)

	// 测试Set新对象
	newPb := &MockPbData{Value: 300}
	wrapper := link.Set(1, newPb)

	if wrapper == nil {
		t.Fatal("Set returned nil wrapper")
	}
	if wrapper.pb.Value != 300 {
		t.Errorf("Expected value 300, got %d", wrapper.pb.Value)
	}
	if !wrapper.IsLinked() {
		t.Error("Expected wrapper to be linked")
	}

	// 验证protobuf map被更新
	if pbMap[1].Value != 300 {
		t.Errorf("Expected pbMap[1].Value to be 300, got %d", pbMap[1].Value)
	}

	// 验证脏标记
	if !tracker.IsDirty(dirtyBit) {
		t.Error("Expected tracker to be marked dirty")
	}

	// 测试Set替换现有对象
	tracker.ClearAllDirty() // 清除脏标记
	oldWrapper := wrapper

	newPb2 := &MockPbData{Value: 400}
	wrapper2 := link.Set(1, newPb2)

	if wrapper2 == nil {
		t.Fatal("Set returned nil wrapper")
	}
	if wrapper2.pb.Value != 400 {
		t.Errorf("Expected value 400, got %d", wrapper2.pb.Value)
	}

	// 验证旧对象被Unlink
	if oldWrapper.IsLinked() {
		t.Error("Expected old wrapper to be unlinked")
	}

	// 验证新对象被Link
	if !wrapper2.IsLinked() {
		t.Error("Expected new wrapper to be linked")
	}

	// 验证脏标记再次被设置
	if !tracker.IsDirty(dirtyBit) {
		t.Error("Expected tracker to be marked dirty after replace")
	}
}

func TestXMapLink_Delete(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}

	tracker := &dt.DirtyTracker{}
	const dirtyBit = int64(1 << 0)

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		dirtyBit,
		NewMockWrapper,
		true,
	)

	// 获取包装对象的引用
	wrapper, _ := link.Get(1)

	// 清除脏标记
	tracker.ClearAllDirty()

	// 测试删除存在的key
	deleted := link.Delete(1)
	if !deleted {
		t.Error("Expected Delete to return true")
	}

	// 验证对象被Unlink
	if wrapper.IsLinked() {
		t.Error("Expected wrapper to be unlinked after delete")
	}

	// 验证map被更新
	if link.Has(1) {
		t.Error("Expected key 1 to be deleted")
	}
	if _, ok := pbMap[1]; ok {
		t.Error("Expected pbMap[1] to be deleted")
	}

	// 验证脏标记
	if !tracker.IsDirty(dirtyBit) {
		t.Error("Expected tracker to be marked dirty after delete")
	}

	// 测试删除不存在的key
	tracker.ClearAllDirty()
	deleted = link.Delete(999)
	if deleted {
		t.Error("Expected Delete to return false for non-existent key")
	}

	// 删除不存在的key不应该标记脏位
	if tracker.IsDirty(dirtyBit) {
		t.Error("Expected tracker to not be marked dirty when deleting non-existent key")
	}
}

func TestXMapLink_Has(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}

	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 测试存在的key
	if !link.Has(1) {
		t.Error("Expected key 1 to exist")
	}

	// 测试不存在的key
	if link.Has(999) {
		t.Error("Expected key 999 to not exist")
	}
}

func TestXMapLink_Len(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 测试空map
	if link.Len() != 0 {
		t.Errorf("Expected length 0, got %d", link.Len())
	}

	// 添加元素
	link.Set(1, &MockPbData{Value: 100})
	if link.Len() != 1 {
		t.Errorf("Expected length 1, got %d", link.Len())
	}

	link.Set(2, &MockPbData{Value: 200})
	if link.Len() != 2 {
		t.Errorf("Expected length 2, got %d", link.Len())
	}

	// 删除元素
	link.Delete(1)
	if link.Len() != 1 {
		t.Errorf("Expected length 1, got %d", link.Len())
	}
}

func TestXMapLink_Range(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}
	pbMap[2] = &MockPbData{Value: 200}
	pbMap[3] = &MockPbData{Value: 300}

	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 测试遍历所有元素
	count := 0
	sum := int32(0)
	link.Range(func(key int64, wrapper *MockWrapper) bool {
		count++
		sum += wrapper.pb.Value
		return true
	})

	if count != 3 {
		t.Errorf("Expected to iterate 3 times, got %d", count)
	}
	if sum != 600 {
		t.Errorf("Expected sum 600, got %d", sum)
	}

	// 测试提前退出
	count = 0
	link.Range(func(key int64, wrapper *MockWrapper) bool {
		count++
		return count < 2 // 只遍历前2个
	})

	if count != 2 {
		t.Errorf("Expected to iterate 2 times, got %d", count)
	}
}

func TestXMapLink_Clear(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}
	pbMap[2] = &MockPbData{Value: 200}

	tracker := &dt.DirtyTracker{}
	const dirtyBit = int64(1 << 0)

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		dirtyBit,
		NewMockWrapper,
		true,
	)

	// 获取包装对象的引用
	wrapper1, _ := link.Get(1)
	wrapper2, _ := link.Get(2)

	// 清除脏标记
	tracker.ClearAllDirty()

	// 测试Clear
	link.Clear()

	// 验证长度为0
	if link.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", link.Len())
	}

	// 验证所有对象被Unlink
	if wrapper1.IsLinked() {
		t.Error("Expected wrapper1 to be unlinked after clear")
	}
	if wrapper2.IsLinked() {
		t.Error("Expected wrapper2 to be unlinked after clear")
	}

	// 验证protobuf map被清空
	if len(pbMap) != 0 {
		t.Errorf("Expected pbMap to be empty, got length %d", len(pbMap))
	}

	// 验证脏标记
	if !tracker.IsDirty(dirtyBit) {
		t.Error("Expected tracker to be marked dirty after clear")
	}
}

func TestXMapLink_Keys(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}
	pbMap[2] = &MockPbData{Value: 200}
	pbMap[3] = &MockPbData{Value: 300}

	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	keys := link.Keys()

	// 验证key数量
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// 验证所有key都存在
	keyMap := make(map[int64]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	if !keyMap[1] || !keyMap[2] || !keyMap[3] {
		t.Error("Missing expected keys")
	}
}

func TestXMapLink_Values(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}
	pbMap[2] = &MockPbData{Value: 200}
	pbMap[3] = &MockPbData{Value: 300}

	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	values := link.Values()

	// 验证value数量
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	// 验证所有value都存在
	sum := int32(0)
	for _, wrapper := range values {
		sum += wrapper.pb.Value
	}

	if sum != 600 {
		t.Errorf("Expected sum 600, got %d", sum)
	}
}

func TestXMapLink_DirtyMarkPropagation(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}
	const dirtyBit = int64(1 << 0)

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		dirtyBit,
		NewMockWrapper,
		true,
	)

	// 添加对象
	link.Set(1, &MockPbData{Value: 100})
	tracker.ClearAllDirty()

	// 获取包装对象
	wrapper, _ := link.Get(1)

	// 修改包装对象（标记脏位）
	const childBit = int64(1 << 1)
	wrapper.MarkDirty(childBit)

	// 验证父对象被标记为脏
	if !tracker.IsDirty(dirtyBit) {
		t.Error("Expected parent tracker to be marked dirty after child modification")
	}
}

func TestXMapLink_EmptyMap(t *testing.T) {
	var pbMap map[int64]*MockPbData // nil map
	tracker := &dt.DirtyTracker{}

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 测试操作空map
	if link.Len() != 0 {
		t.Errorf("Expected length 0, got %d", link.Len())
	}

	if link.Has(1) {
		t.Error("Expected key 1 to not exist")
	}

	_, ok := link.Get(1)
	if ok {
		t.Error("Expected Get to return false for empty map")
	}

	// Set应该能够创建map
	link.Set(1, &MockPbData{Value: 100})
	if link.Len() != 1 {
		t.Errorf("Expected length 1 after set, got %d", link.Len())
	}
}

// ============================================
// 基准测试
// ============================================

func BenchmarkXMapLink_Set(b *testing.B) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		link.Set(int64(i%100), &MockPbData{Value: int32(i)})
	}
}

func BenchmarkXMapLink_Get(b *testing.B) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 预填充数据
	for i := 0; i < 100; i++ {
		link.Set(int64(i), &MockPbData{Value: int32(i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		link.Get(int64(i % 100))
	}
}

func BenchmarkXMapLink_Delete(b *testing.B) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 添加和删除交替进行
		if i%2 == 0 {
			link.Set(int64(i%100), &MockPbData{Value: int32(i)})
		} else {
			link.Delete(int64((i - 1) % 100))
		}
	}
}

func BenchmarkXMapLink_Range(b *testing.B) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 预填充数据
	for i := 0; i < 100; i++ {
		link.Set(int64(i), &MockPbData{Value: int32(i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		link.Range(func(key int64, wrapper *MockWrapper) bool {
			return true
		})
	}
}
