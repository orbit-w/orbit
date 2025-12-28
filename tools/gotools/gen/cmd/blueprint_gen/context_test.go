package blueprint_gen

import (
	"testing"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/mmeobj"
	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/pb_gen/net_message"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// TestBlueprintContext_BuildSymbolTable 测试符号表构建
func TestBlueprintContext_BuildSymbolTable(t *testing.T) {
	ctx := NewBlueprintContext()

	// 创建测试数据
	entity := mmeobj.NewEntity()
	entity.Name = "PlayerEntity"
	entity.SourceFile = "player.yaml"
	ctx.AddEntity(entity)

	manager := mmeobj.NewManager()
	manager.Name = "HeroManager"
	manager.SourceFile = "hero.yaml"
	ctx.AddManager(manager)

	// 构建符号表
	err := ctx.BuildSymbolTable()
	if err != nil {
		t.Fatalf("BuildSymbolTable() failed: %v", err)
	}

	// 验证符号表
	if ctx.SymbolTable == nil {
		t.Fatal("SymbolTable is nil")
	}

	// 验证符号数量
	expectedSize := 2 // PlayerEntity + HeroManager
	if ctx.SymbolTable.Size() != expectedSize {
		t.Errorf("SymbolTable size = %d, want %d", ctx.SymbolTable.Size(), expectedSize)
	}

	// 验证可以查找到符号
	if !ctx.SymbolTable.Has("PlayerEntity") {
		t.Error("PlayerEntity not found in symbol table")
	}
	if !ctx.SymbolTable.Has("HeroManager") {
		t.Error("HeroManager not found in symbol table")
	}

	// 验证带作用域的查找
	if !ctx.SymbolTable.Has("mme.PlayerEntity") {
		t.Error("mme.PlayerEntity not found in symbol table")
	}
}

// TestBlueprintContext_ResolveReferences 测试类型引用解析
func TestBlueprintContext_ResolveReferences(t *testing.T) {
	ctx := NewBlueprintContext()

	// 创建 HeroManager
	heroManager := mmeobj.NewManager()
	heroManager.Name = "HeroManager"
	heroManager.SourceFile = "hero.yaml"
	heroManager.AddField(&types.Field{
		Name:   "level",
		Number: 1,
		Type: types.FieldType{
			Kind: types.FieldKindInt32,
			Name: "int32",
		},
	})
	ctx.AddManager(heroManager)

	// 创建 PlayerEntity，引用 HeroManager
	entity := mmeobj.NewEntity()
	entity.Name = "PlayerEntity"
	entity.SourceFile = "player.yaml"
	entity.AddField(&types.Field{
		Name:   "heroes",
		Number: 1,
		Type: types.FieldType{
			Kind:     types.FieldKindMMEObject,
			TypeName: "HeroManager",
		},
	})
	ctx.AddEntity(entity)

	// 构建符号表
	if err := ctx.BuildSymbolTable(); err != nil {
		t.Fatalf("BuildSymbolTable() failed: %v", err)
	}

	// 解析引用
	err := ctx.ResolveReferences()
	if err != nil {
		t.Fatalf("ResolveReferences() failed: %v", err)
	}

	// 验证引用已解析
	heroesField := entity.Fields[0]
	if !heroesField.Type.IsResolved() {
		t.Error("heroes field type is not resolved")
	}

	// 验证解析结果
	def := heroesField.Type.GetResolvedDefinition()
	if def == nil {
		t.Fatal("ResolvedDefinition is nil")
	}
	if def.GetName() != "HeroManager" {
		t.Errorf("ResolvedDefinition name = %s, want HeroManager", def.GetName())
	}
	if def.GetKind() != types.DefKindManager {
		t.Errorf("ResolvedDefinition kind = %v, want DefKindManager", def.GetKind())
	}
}

// TestBlueprintContext_ResolveReferences_UndefinedType 测试未定义类型的错误处理
func TestBlueprintContext_ResolveReferences_UndefinedType(t *testing.T) {
	ctx := NewBlueprintContext()

	// 创建 Entity，引用不存在的 ItemManager
	entity := mmeobj.NewEntity()
	entity.Name = "PlayerEntity"
	entity.SourceFile = "player.yaml"
	entity.AddField(&types.Field{
		Name:   "items",
		Number: 1,
		Type: types.FieldType{
			Kind:     types.FieldKindMMEObject,
			TypeName: "ItemManager", // 未定义
		},
	})
	ctx.AddEntity(entity)

	// 构建符号表
	if err := ctx.BuildSymbolTable(); err != nil {
		t.Fatalf("BuildSymbolTable() failed: %v", err)
	}

	// 解析引用（应该失败）
	err := ctx.ResolveReferences()
	if err == nil {
		t.Error("ResolveReferences() should fail for undefined type")
	}
}

// TestBlueprintContext_ResolveNestedTypes 测试嵌套类型的解析
func TestBlueprintContext_ResolveNestedTypes(t *testing.T) {
	ctx := NewBlueprintContext()

	// 创建 HeroManager
	heroManager := mmeobj.NewManager()
	heroManager.Name = "HeroManager"
	heroManager.SourceFile = "hero.yaml"
	ctx.AddManager(heroManager)

	// 创建 PlayerEntity，使用 map<int32, HeroManager>
	entity := mmeobj.NewEntity()
	entity.Name = "PlayerEntity"
	entity.SourceFile = "player.yaml"
	entity.AddField(&types.Field{
		Name:   "heroMap",
		Number: 1,
		Type: types.FieldType{
			Kind: types.FieldKindMap,
			KeyType: &types.FieldType{
				Kind: types.FieldKindInt32,
				Name: "int32",
			},
			ValueType: &types.FieldType{
				Kind:     types.FieldKindMMEObject,
				TypeName: "HeroManager",
			},
		},
	})
	ctx.AddEntity(entity)

	// 构建符号表
	if err := ctx.BuildSymbolTable(); err != nil {
		t.Fatalf("BuildSymbolTable() failed: %v", err)
	}

	// 解析引用
	err := ctx.ResolveReferences()
	if err != nil {
		t.Fatalf("ResolveReferences() failed: %v", err)
	}

	// 验证嵌套类型的引用已解析
	heroMapField := entity.Fields[0]
	if !heroMapField.Type.ValueType.IsResolved() {
		t.Error("map value type is not resolved")
	}

	// 验证解析结果
	def := heroMapField.Type.ValueType.GetResolvedDefinition()
	if def == nil {
		t.Fatal("ResolvedDefinition is nil")
	}
	if def.GetName() != "HeroManager" {
		t.Errorf("ResolvedDefinition name = %s, want HeroManager", def.GetName())
	}
}

// TestBlueprintContext_NetWallMessages 测试 NetWall 消息的符号表注册和引用解析
func TestBlueprintContext_NetWallMessages(t *testing.T) {
	ctx := NewBlueprintContext()

	// 创建 NetWall 文件
	netwall := NewNetWallFile()
	netwall.Name = "core"
	netwall.PackageName = "Core"

	// 创建 Request 消息
	request := net_message.NewNetMessage(net_message.NetWallMessageTypeRequest, "Login")
	request.PackageName = "Core"
	request.SourceFile = "core.yaml"
	netwall.AddRequest(request)

	// 创建 DataStruct 消息
	dataStruct := net_message.NewNetMessage(net_message.NetWallMessageTypeDataStruct, "PlayerInfo")
	dataStruct.PackageName = "Core"
	dataStruct.SourceFile = "core.yaml"
	netwall.AddDataStruct(dataStruct)

	ctx.AddNetWallFile(netwall)

	// 构建符号表
	if err := ctx.BuildSymbolTable(); err != nil {
		t.Fatalf("BuildSymbolTable() failed: %v", err)
	}

	// 验证 NetWall 消息已注册
	if !ctx.SymbolTable.Has("Request_Login") {
		t.Error("Request_Login not found in symbol table")
	}
	if !ctx.SymbolTable.Has("PlayerInfo") {
		t.Error("PlayerInfo not found in symbol table")
	}

	// 验证带作用域的查找
	if !ctx.SymbolTable.Has("Core.Request_Login") {
		t.Error("Core.Request_Login not found in symbol table")
	}
}

// TestMMEObject_ImplementsDefinition 测试 MMEObject 实现 Definition 接口
func TestMMEObject_ImplementsDefinition(t *testing.T) {
	manager := mmeobj.NewManager()
	manager.Name = "HeroManager"
	manager.SourceFile = "hero.yaml"
	manager.AddField(&types.Field{
		Name:   "level",
		Number: 1,
		Type: types.FieldType{
			Kind: types.FieldKindInt32,
			Name: "int32",
		},
	})

	// 验证实现了 Definition 接口
	var def types.Definition = manager.MMEObject

	if def.GetName() != "HeroManager" {
		t.Errorf("GetName() = %s, want HeroManager", def.GetName())
	}
	if def.GetKind() != types.DefKindManager {
		t.Errorf("GetKind() = %v, want DefKindManager", def.GetKind())
	}
	if def.GetFile() != "hero.yaml" {
		t.Errorf("GetFile() = %s, want hero.yaml", def.GetFile())
	}
	if len(def.GetFields()) != 1 {
		t.Errorf("GetFields() length = %d, want 1", len(def.GetFields()))
	}
}

// TestNetMessage_ImplementsDefinition 测试 NetMessage 实现 Definition 接口
func TestNetMessage_ImplementsDefinition(t *testing.T) {
	message := net_message.NewNetMessage(net_message.NetWallMessageTypeRequest, "Login")
	message.SourceFile = "core.yaml"
	message.SetFields([]*types.Field{
		{
			Name:   "username",
			Number: 1,
			Type: types.FieldType{
				Kind: types.FieldKindString,
				Name: "string",
			},
		},
	})

	// 验证实现了 Definition 接口
	var def types.Definition = message

	// 对于 Request 类型，GetName() 返回 FullName (Request_Login)
	expectedName := "Request_Login"
	if def.GetName() != expectedName {
		t.Errorf("GetName() = %s, want %s", def.GetName(), expectedName)
	}
	if def.GetKind() != types.DefKindMessage {
		t.Errorf("GetKind() = %v, want DefKindMessage", def.GetKind())
	}
	if def.GetFile() != "core.yaml" {
		t.Errorf("GetFile() = %s, want core.yaml", def.GetFile())
	}
	if len(def.GetFields()) != 1 {
		t.Errorf("GetFields() length = %d, want 1", len(def.GetFields()))
	}
}

