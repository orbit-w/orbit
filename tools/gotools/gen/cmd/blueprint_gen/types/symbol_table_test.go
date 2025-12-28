package blueprint_types

import (
	"testing"
)

// mockDefinition 用于测试的 Definition 实现
type mockDefinition struct {
	name   string
	kind   DefinitionKind
	file   string
	fields []*Field
}

func (m *mockDefinition) GetName() string {
	return m.name
}

func (m *mockDefinition) GetKind() DefinitionKind {
	return m.kind
}

func (m *mockDefinition) GetFile() string {
	return m.file
}

func (m *mockDefinition) GetFields() []*Field {
	return m.fields
}

func newMockDefinition(name string, kind DefinitionKind, file string) *mockDefinition {
	return &mockDefinition{
		name:   name,
		kind:   kind,
		file:   file,
		fields: make([]*Field, 0),
	}
}

func TestSymbolTable_Register(t *testing.T) {
	st := NewSymbolTable()

	entity := newMockDefinition("PlayerEntity", DefKindEntity, "player.yaml")
	err := st.Register(entity)
	if err != nil {
		t.Errorf("Register() failed: %v", err)
	}

	if st.Size() != 1 {
		t.Errorf("SymbolTable size = %d, want 1", st.Size())
	}
}

func TestSymbolTable_Register_Duplicate(t *testing.T) {
	st := NewSymbolTable()

	entity1 := newMockDefinition("PlayerEntity", DefKindEntity, "player.yaml")
	entity2 := newMockDefinition("PlayerEntity", DefKindEntity, "player2.yaml")

	err1 := st.Register(entity1)
	if err1 != nil {
		t.Errorf("First Register() failed: %v", err1)
	}

	err2 := st.Register(entity2)
	if err2 == nil {
		t.Errorf("Register() should fail for duplicate symbol")
	}
}

func TestSymbolTable_Lookup(t *testing.T) {
	st := NewSymbolTable()

	entity := newMockDefinition("PlayerEntity", DefKindEntity, "player.yaml")
	_ = st.Register(entity)

	def, exists := st.Lookup("PlayerEntity")
	if !exists {
		t.Errorf("Lookup() should find 'PlayerEntity'")
	}
	if def.GetName() != "PlayerEntity" {
		t.Errorf("Lookup() returned wrong definition, got %s", def.GetName())
	}

	_, exists = st.Lookup("NonExistent")
	if exists {
		t.Errorf("Lookup() should not find 'NonExistent'")
	}
}

func TestSymbolTable_RegisterWithScope(t *testing.T) {
	st := NewSymbolTable()

	entity := newMockDefinition("PlayerEntity", DefKindEntity, "player.yaml")
	err := st.RegisterWithScope("mme", entity)
	if err != nil {
		t.Errorf("RegisterWithScope() failed: %v", err)
	}

	// 测试简单名称查找
	def, exists := st.Lookup("PlayerEntity")
	if !exists {
		t.Errorf("Lookup() should find 'PlayerEntity'")
	}

	// 测试带作用域的名称查找
	def, exists = st.Lookup("mme.PlayerEntity")
	if !exists {
		t.Errorf("Lookup() should find 'mme.PlayerEntity'")
	}
	if def.GetName() != "PlayerEntity" {
		t.Errorf("Lookup() returned wrong definition")
	}
}

func TestReferenceResolver_ResolveField(t *testing.T) {
	st := NewSymbolTable()

	// 注册类型定义
	heroManager := newMockDefinition("HeroManager", DefKindManager, "hero.yaml")
	_ = st.Register(heroManager)

	// 创建解析器
	resolver := NewReferenceResolver(st)

	// 创建字段
	field := &Field{
		Name:   "heroes",
		Number: 1,
	}
	field.Type.Kind = FieldKindMMEObject
	field.Type.TypeName = "HeroManager"

	// 解析字段
	err := resolver.ResolveField(field)
	if err != nil {
		t.Errorf("ResolveField() failed: %v", err)
	}

	// 检查是否已解析
	if !field.Type.IsResolved() {
		t.Errorf("Field type should be resolved")
	}

	// 检查解析结果
	def := field.Type.GetResolvedDefinition()
	if def == nil {
		t.Errorf("ResolvedType should not be nil")
	}
	if def.GetName() != "HeroManager" {
		t.Errorf("ResolvedType name = %s, want HeroManager", def.GetName())
	}
}

func TestReferenceResolver_ResolveField_Undefined(t *testing.T) {
	st := NewSymbolTable()
	resolver := NewReferenceResolver(st)

	// 创建引用未定义类型的字段
	field := &Field{
		Name:   "items",
		Number: 1,
	}
	field.Type.Kind = FieldKindMessage
	field.Type.TypeName = "ItemManager"

	// 解析字段（应该失败）
	err := resolver.ResolveField(field)
	if err == nil {
		t.Errorf("ResolveField() should fail for undefined type")
	}

	// 检查未解析
	if field.Type.IsResolved() {
		t.Errorf("Field type should not be resolved")
	}
}

func TestReferenceResolver_ResolveFields(t *testing.T) {
	st := NewSymbolTable()

	// 注册类型定义
	heroManager := newMockDefinition("HeroManager", DefKindManager, "hero.yaml")
	playerEntity := newMockDefinition("PlayerEntity", DefKindEntity, "player.yaml")
	_ = st.Register(heroManager)
	_ = st.Register(playerEntity)

	// 创建解析器
	resolver := NewReferenceResolver(st)

	// 创建字段列表
	fields := []*Field{
		{
			Name:   "heroes",
			Number: 1,
			Type: FieldType{
				Kind:     FieldKindMMEObject,
				TypeName: "HeroManager",
			},
		},
		{
			Name:   "player",
			Number: 2,
			Type: FieldType{
				Kind:     FieldKindMMEObject,
				TypeName: "PlayerEntity",
			},
		},
		{
			Name:   "items",
			Number: 3,
			Type: FieldType{
				Kind:     FieldKindMessage,
				TypeName: "ItemManager", // 未定义
			},
		},
	}

	// 批量解析
	errors := resolver.ResolveFields(fields)

	// 检查错误（应该有一个错误：ItemManager 未定义）
	if len(errors) != 1 {
		t.Errorf("ResolveFields() should have 1 error, got %d", len(errors))
	}

	// 检查第一个字段已解析
	if !fields[0].Type.IsResolved() {
		t.Errorf("Field 'heroes' should be resolved")
	}

	// 检查第二个字段已解析
	if !fields[1].Type.IsResolved() {
		t.Errorf("Field 'player' should be resolved")
	}

	// 检查第三个字段未解析
	if fields[2].Type.IsResolved() {
		t.Errorf("Field 'items' should not be resolved")
	}
}

func TestFieldType_IsResolved(t *testing.T) {
	tests := []struct {
		name string
		ft   FieldType
		want bool
	}{
		{
			name: "resolved MMEObject",
			ft: FieldType{
				Kind:         FieldKindMMEObject,
				ResolvedType: newMockDefinition("HeroManager", DefKindManager, "hero.yaml"),
			},
			want: true,
		},
		{
			name: "unresolved MMEObject",
			ft: FieldType{
				Kind:         FieldKindMMEObject,
				ResolvedType: nil,
			},
			want: false,
		},
		{
			name: "base type",
			ft: FieldType{
				Kind:         FieldKindInt32,
				ResolvedType: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ft.IsResolved(); got != tt.want {
				t.Errorf("FieldType.IsResolved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefinitionKind_String(t *testing.T) {
	tests := []struct {
		kind DefinitionKind
		want string
	}{
		{DefKindEntity, "Entity"},
		{DefKindManager, "Manager"},
		{DefKindModule, "Module"},
		{DefKindMechanism, "Mechanism"},
		{DefKindMessage, "Message"},
		{DefKindEnum, "Enum"},
		{DefKindStruct, "Struct"},
		{DefKindUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("DefinitionKind.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReferenceResolver_ResolveNestedTypes(t *testing.T) {
	st := NewSymbolTable()

	// 注册类型定义
	heroManager := newMockDefinition("HeroManager", DefKindManager, "hero.yaml")
	_ = st.Register(heroManager)

	// 创建解析器
	resolver := NewReferenceResolver(st)

	// 创建嵌套类型字段: map<int32, HeroManager>
	valueType := &FieldType{
		Kind:     FieldKindMMEObject,
		TypeName: "HeroManager",
	}
	keyType := &FieldType{
		Kind: FieldKindInt32,
		Name: "int32",
	}

	field := &Field{
		Name:   "heroMap",
		Number: 1,
		Type: FieldType{
			Kind:      FieldKindMap,
			KeyType:   keyType,
			ValueType: valueType,
		},
	}

	// 解析字段
	err := resolver.ResolveField(field)
	if err != nil {
		t.Errorf("ResolveField() failed: %v", err)
	}

	// 检查嵌套的 Value 类型是否已解析
	if !field.Type.ValueType.IsResolved() {
		t.Errorf("Nested value type should be resolved")
	}

	// 检查解析结果
	def := field.Type.ValueType.GetResolvedDefinition()
	if def == nil {
		t.Errorf("Nested ResolvedType should not be nil")
	}
	if def.GetName() != "HeroManager" {
		t.Errorf("Nested ResolvedType name = %s, want HeroManager", def.GetName())
	}
}

