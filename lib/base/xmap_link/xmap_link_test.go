package xmaplink

import (
	"testing"

	dt "gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
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
	pb            *MockPbData
	tracker       dt.DirtyTracker
	factsAccessor func() // 修正：factsAccessor 是 func() 类型，不是接口
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

func (w *MockWrapper) LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64, factsAccessor func()) {
	w.factsAccessor = factsAccessor
	w.tracker.Link(parent, parentBit)
}

func (w *MockWrapper) Unlink() {
	w.tracker.Unlink()
	w.factsAccessor = nil
}

func (w *MockWrapper) GetDirtyTracker() *dt.DirtyTracker {
	return &w.tracker
}

func (w *MockWrapper) MarkDirty(dirtyBit int64) {
	w.tracker.MarkDirty(dirtyBit)
	if w.factsAccessor != nil {
		w.factsAccessor()
	}
}

func (w *MockWrapper) IsLinked() bool {
	return w.factsAccessor != nil
}

func (w *MockWrapper) GetValue() int32 {
	return w.pb.Value
}

func (w *MockWrapper) SetValue(value int32) {
	w.pb.Value = value
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

// ============================================
// 新增的高级测试用例
// ============================================

// TestXMapLink_SetParent 测试动态设置父节点
func TestXMapLink_SetParent(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}
	pbMap[2] = &MockPbData{Value: 200}

	// 创建第一个父tracker
	tracker1 := &dt.DirtyTracker{}
	const dirtyBit1 = int64(1 << 0)

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker1,
		dirtyBit1,
		NewMockWrapper,
		true,
	)

	// 验证初始状态：所有wrapper都链接到tracker1
	wrapper1, _ := link.Get(1)
	if !wrapper1.IsLinked() {
		t.Error("Expected wrapper1 to be linked initially")
	}

	// 创建第二个父tracker并重新设置
	tracker2 := &dt.DirtyTracker{}
	const dirtyBit2 = int64(1 << 2)

	// 调用SetParent应该重新链接所有现有的wrapper
	link.SetParent(tracker2, dirtyBit2)

	// SetParent后，wrapper仍然应该是链接状态
	if !wrapper1.IsLinked() {
		t.Error("Expected wrapper1 to still be linked after SetParent")
	}

	// 清除所有脏标记
	tracker1.ClearAllDirty()
	tracker2.ClearAllDirty()

	// 重新获取wrapper（确保获取最新的链接状态）
	wrapper1, _ = link.Get(1)

	// 修改wrapper，应该标记tracker2
	wrapper1.MarkDirty(1 << 1)
	if !tracker2.IsDirty(dirtyBit2) {
		t.Error("Expected tracker2 to be marked dirty after SetParent")
	}

	// 验证添加新元素也使用新的父tracker
	tracker2.ClearAllDirty()
	wrapper3 := link.Set(3, &MockPbData{Value: 300})

	// 新添加的wrapper应该链接到tracker2
	if !wrapper3.IsLinked() {
		t.Error("Expected new wrapper to be linked")
	}

	tracker2.ClearAllDirty()
	wrapper3.MarkDirty(1 << 1)
	if !tracker2.IsDirty(dirtyBit2) {
		t.Error("Expected tracker2 to be marked dirty for new wrapper")
	}
}

// TestXMapLink_GetMapAccessor 测试获取MapAccessor
func TestXMapLink_GetMapAccessor(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	accessor := link.GetMapAccessor()
	if accessor == nil {
		t.Fatal("GetMapAccessor returned nil")
	}

	// 通过accessor直接操作应该也能工作
	pb := &MockPbData{Value: 999}
	accessor.Set(1, pb)

	// 验证pbMap被更新
	if pbMap[1].Value != 999 {
		t.Errorf("Expected pbMap[1].Value to be 999, got %d", pbMap[1].Value)
	}
}

// TestXMapLink_RangeWithNilCallback 测试Range传入nil回调
func TestXMapLink_RangeWithNilCallback(t *testing.T) {
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

	// 传入nil不应该panic
	link.Range(nil)
}

// TestXMapLink_FactsAccessorPropagation 测试FactsAccessor的正确传播
func TestXMapLink_FactsAccessorPropagation(t *testing.T) {
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

	// 添加一个对象
	wrapper := link.Set(1, &MockPbData{Value: 100})
	tracker.ClearAllDirty()

	// 调用wrapper的MarkDirty应该通过factsAccessor触发TrackSet
	wrapper.MarkDirty(1 << 1)

	// 验证父tracker被标记为脏（通过factsAccessor）
	if !tracker.IsDirty(dirtyBit) {
		t.Error("Expected parent tracker to be marked dirty via factsAccessor")
	}
}

// TestXMapLink_MultipleSetSameKey 测试同一个key多次Set
func TestXMapLink_MultipleSetSameKey(t *testing.T) {
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

	// 第一次Set
	wrapper1 := link.Set(1, &MockPbData{Value: 100})
	if !wrapper1.IsLinked() {
		t.Error("Expected wrapper1 to be linked")
	}

	// 第二次Set相同的key
	wrapper2 := link.Set(1, &MockPbData{Value: 200})

	// 验证wrapper1被unlink
	if wrapper1.IsLinked() {
		t.Error("Expected wrapper1 to be unlinked after replacement")
	}

	// 验证wrapper2被link
	if !wrapper2.IsLinked() {
		t.Error("Expected wrapper2 to be linked")
	}

	// 验证map长度仍然为1
	if link.Len() != 1 {
		t.Errorf("Expected length 1, got %d", link.Len())
	}

	// 验证新值被正确设置
	currentWrapper, _ := link.Get(1)
	if currentWrapper.GetValue() != 200 {
		t.Errorf("Expected value 200, got %d", currentWrapper.GetValue())
	}

	// 第三次Set
	wrapper3 := link.Set(1, &MockPbData{Value: 300})

	// 验证wrapper2被unlink
	if wrapper2.IsLinked() {
		t.Error("Expected wrapper2 to be unlinked after second replacement")
	}

	// 验证wrapper3被link
	if !wrapper3.IsLinked() {
		t.Error("Expected wrapper3 to be linked")
	}
}

// TestXMapLink_ClearThenAdd 测试清空后再添加
func TestXMapLink_ClearThenAdd(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	pbMap[1] = &MockPbData{Value: 100}
	pbMap[2] = &MockPbData{Value: 200}

	tracker := &dt.DirtyTracker{}
	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 清空
	link.Clear()
	if link.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", link.Len())
	}

	// 再次添加
	wrapper := link.Set(3, &MockPbData{Value: 300})
	if !wrapper.IsLinked() {
		t.Error("Expected new wrapper to be linked")
	}

	if link.Len() != 1 {
		t.Errorf("Expected length 1, got %d", link.Len())
	}

	// 验证新数据正确
	w, ok := link.Get(3)
	if !ok {
		t.Error("Expected key 3 to exist")
	}
	if w.GetValue() != 300 {
		t.Errorf("Expected value 300, got %d", w.GetValue())
	}
}

// TestXMapLink_OperationsOnNilMap 测试在nil map上的操作
func TestXMapLink_OperationsOnNilMap(t *testing.T) {
	var pbMap map[int64]*MockPbData // nil map
	tracker := &dt.DirtyTracker{}

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 在nil map上Set应该能够创建map并正常工作
	wrapper := link.Set(1, &MockPbData{Value: 100})

	if !wrapper.IsLinked() {
		t.Error("Expected wrapper to be linked")
	}

	if link.Len() != 1 {
		t.Errorf("Expected length 1, got %d", link.Len())
	}

	// 验证可以Get
	w, ok := link.Get(1)
	if !ok {
		t.Error("Expected key 1 to exist")
	}
	if w.GetValue() != 100 {
		t.Errorf("Expected value 100, got %d", w.GetValue())
	}
}

// TestXMapLink_KeysAndValues 测试Keys和Values的顺序无关性
func TestXMapLink_KeysAndValues(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 添加多个元素
	expectedKeys := []int64{1, 2, 3, 5, 8, 13}
	expectedSum := int32(0)
	for _, key := range expectedKeys {
		value := int32(key * 10)
		link.Set(key, &MockPbData{Value: value})
		expectedSum += value
	}

	// 测试Keys
	keys := link.Keys()
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	keyMap := make(map[int64]bool)
	for _, key := range keys {
		keyMap[key] = true
	}
	for _, expected := range expectedKeys {
		if !keyMap[expected] {
			t.Errorf("Expected key %d not found", expected)
		}
	}

	// 测试Values
	values := link.Values()
	if len(values) != len(expectedKeys) {
		t.Errorf("Expected %d values, got %d", len(expectedKeys), len(values))
	}

	actualSum := int32(0)
	for _, wrapper := range values {
		actualSum += wrapper.GetValue()
	}
	if actualSum != expectedSum {
		t.Errorf("Expected sum %d, got %d", expectedSum, actualSum)
	}
}

// TestXMapLink_EmptyOperations 测试空map上的所有操作
func TestXMapLink_EmptyOperations(t *testing.T) {
	pbMap := make(map[int64]*MockPbData)
	tracker := &dt.DirtyTracker{}

	link := NewXMapLinkWithParent[int64, *MockPbData, *MockWrapper](
		&pbMap,
		tracker,
		1<<0,
		NewMockWrapper,
		true,
	)

	// 测试所有操作在空map上都不会panic
	if link.Len() != 0 {
		t.Error("Expected empty map to have length 0")
	}

	if link.Has(1) {
		t.Error("Expected Has to return false on empty map")
	}

	_, ok := link.Get(1)
	if ok {
		t.Error("Expected Get to return false on empty map")
	}

	if link.Delete(1) {
		t.Error("Expected Delete to return false on empty map")
	}

	// Range在空map上应该不执行回调
	executed := false
	link.Range(func(key int64, wrapper *MockWrapper) bool {
		executed = true
		return true
	})
	if executed {
		t.Error("Expected Range callback not to be executed on empty map")
	}

	// Clear空map应该不会panic
	link.Clear()

	// Keys和Values应该返回空切片
	keys := link.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected empty keys slice, got length %d", len(keys))
	}

	values := link.Values()
	if len(values) != 0 {
		t.Errorf("Expected empty values slice, got length %d", len(values))
	}
}
