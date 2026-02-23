# Blueprint Gen — 完整示例集

## 示例一：添加"英雄"模块（完整流程）

### 目标
在 PlayerEntity 下新增一个 HeroManager，管理多个英雄，每个英雄包含基础数据和升级数据。

### Step 1：mechanisms.yaml

```yaml
Mechanisms:
  - HeroMechanism:                     # 英雄基础数据
      HeroMechanism:
        int32 HeroId: 1
        string HeroName: 2
        int32 Level: 3
        int32 Exp: 4

  - LevelUpMechanism:                  # 升级记录
      LevelUpMechanism:
        int32 TotalLevelUpCount: 1
        int64 LastLevelUpTime: 2

Register:
  - HeroMechanism: 1
  - LevelUpMechanism: 2
```

### Step 2：modules.yaml

```yaml
Modules:
  - HeroModule:                        # 单个英雄的模块
      - HeroMechanism HeroData: 1
      - LevelUpMechanism LevelData: 2
```

### Step 3：manager.yaml

```yaml
Managers:
  - HeroManager:
      xmap<int32, HeroModule> Heroes: 1   # HeroId -> HeroModule，支持脏标记
```

### Step 4：entities.yaml

```yaml
Entity:
  - PlayerEntity:
      HeroManager HeroMgr: 1
      BagManager BagMgr: 2             # 假设 BagManager 已存在

Register:
  - PlayerEntity: 1
```

### Step 5：运行生成

```bash
go run ./tools/gotools/gen/main.go blueprintgen \
  --blueprint-dir ../protocol/blueprint \
  --proto-output ../protocol/protocol \
  --go-output internal/game/mme \
  --protocol-ids-output pkg/proto/pb \
  --debug
```

### Step 6：生成产物

```
# Proto 文件
protocol/protocol/mechanisms.proto     → HeroMechanism, LevelUpMechanism 消息
protocol/protocol/modules.proto        → HeroModule 消息
protocol/protocol/managers.proto       → HeroManager 消息
protocol/protocol/entities.proto       → PlayerEntity 消息
protocol/protocol/common.proto         → MechanismType, EntityType 枚举

# Go Wrapper
internal/game/mme/mechanisms/hero_mechanism/hero_mechanism_wrapper.go
internal/game/mme/mechanisms/levelup_mechanism/levelup_mechanism_wrapper.go
internal/game/mme/modules/hero_module/hero_module_wrapper.go
internal/game/mme/managers/hero_manager/hero_manager_wrapper.go
internal/game/mme/entities/player/player_entity_wrapper.go
```

---

## 示例二：添加网络消息（NetWall）

### netwall/hero.yaml

```yaml
NetWall:
  Name: Hero
  Requests:
    - GetHeroInfo:
        int32 HeroId: 1
        Response:
          HeroData HeroInfo: 1          # HeroData 引用 headfile.yaml 中的 DataStruct
          int32 Level: 2

    - UpgradeHero:
        int32 HeroId: 1
        int32 TargetLevel: 2
        Response:
          bool Success: 1
          int32 NewLevel: 2

  Notifies:
    - HeroLevelUpNotify:
        int32 HeroId: 1
        int32 OldLevel: 2
        int32 NewLevel: 3
```

### headfile.yaml（共享 DataStruct）

```yaml
DataStruct:
  - HeroData:
      int32 HeroId: 1
      string HeroName: 2
      int32 Level: 3

Enums:
  - HeroRarity:
      HeroRarityUnknown: '0 [Content:"未知"]'
      HeroRarityN: '1 [Content:"普通"]'
      HeroRarityR: '2 [Content:"稀有"]'
      HeroRaritySSR: '3 [Content:"超级稀有"]'
```

### 生成的协议 ID 文件（pkg/proto/pb/）

```go
// protocol_ids.pb.go（自动生成，勿手动编辑）
const (
    ProtocolID_Hero_GetHeroInfo  = 1001
    ProtocolID_Hero_UpgradeHero  = 1002
)
```

---

## 示例三：公共 DataStruct 被 Mechanism 字段引用

```yaml
# headfile.yaml
DataStruct:
  - ItemData:
      int32 ItemId: 1
      int32 Count: 2
      int32 Quality: 3

# mechanisms.yaml
Mechanisms:
  - BagMechanism:
      BagMechanism:
        repeated ItemData Items: 1     # 引用 headfile.yaml 中的 DataStruct
        int32 MaxCapacity: 2
```

> `ItemData` 在 Pass 2 中以 scope `"MME"` 注册，Pass 3 中被成功解析为 `DefKindStruct`。

---

## 示例四：枚举的使用

### mechanisms.yaml 中使用枚举

```yaml
# headfile.yaml 中定义枚举
Enums:
  - HeroRarity:
      HeroRarityUnknown: '0 [Content:"未知"]'
      HeroRarityN: '1 [Content:"普通"]'

# mechanisms.yaml
Mechanisms:
  - HeroMechanism:
      HeroMechanism:
        int32 HeroId: 1
        HeroRarity Rarity: 2           # 使用枚举类型
```

> 枚举字段在 FieldType 中：`Kind = FieldKindEnum`，`TypeName = "HeroRarity"`。

---

## 反例：常见错误示范

### ❌ 错误 1：Mechanism 字段引用 MME 对象

```yaml
# 错误：Mechanism 的字段只能是基础类型
Mechanisms:
  - HeroMechanism:
      HeroMechanism:
        int32 HeroId: 1
        HeroModule SubModule: 2    # ❌ Mechanism 不能引用 Module！
```

**错误信息**：
```
panic: mechanism HeroMechanism field SubModule type is not mme object
```

**正确做法**：如果需要嵌套数据，在 Module 层级组合多个 Mechanism，而不是在 Mechanism 中引用 Module。

---

### ❌ 错误 2：Manager 字段引用 Mechanism（跳过 Module 层）

```yaml
# 错误：Manager 的 Map value 必须是 Module，不能是 Mechanism
Managers:
  - HeroManager:
      xmap<int32, HeroMechanism> Heroes: 1  # ❌ value 必须是 Module！
```

**错误信息**：
```
panic: manager HeroManager map field Heroes value type is not manager
(实际错误：value type is not Module)
```

**正确做法**：先创建 `HeroModule` 包含 `HeroMechanism`，Manager 引用 `HeroModule`。

---

### ❌ 错误 3：字段编号重复或超过 64

```yaml
Mechanisms:
  - HeroMechanism:
      HeroMechanism:
        int32 HeroId: 1
        string HeroName: 1     # ❌ 编号 1 重复！
        int32 Level: 65        # ❌ 超过 ObjectFieldNumberMax = 64！
```

**解决方案**：字段编号需唯一且在 1-64 范围内。

---

### ❌ 错误 4：Register 编号为 0

```yaml
# entities.yaml
Register:
  - PlayerEntity: 0    # ❌ Protobuf 要求第一个枚举值为 0，0 被 Unknown 占用
```

**错误信息**：
```
panic: Register 中 Entity 'PlayerEntity' 的编号不能为 0
```

**解决方案**：Register 编号从 1 开始。

---

### ❌ 错误 5：类型命名不符合命名约定

```yaml
# 错误：类型名不符合 MME 命名约定，会被识别为 Message 而非 MMEObject
Managers:
  - HeroMgr:                  # ❌ 不以 Manager 结尾，无法被 isMMEObjectType 识别
      HeroSub Heroes: 1       # ❌ HeroSub 不以 Module 结尾，识别为 Message
```

**正确做法**：严格遵守命名约定：
- Entity 结尾：`PlayerEntity`
- Manager 结尾：`HeroManager`
- Module 结尾：`HeroModule`
- Mechanism 结尾：`HeroMechanism`

---

### ❌ 错误 6：在 FiledChecker 之前调用导致 panic

```go
// 错误的调用顺序（来自开发经验）
p.ctx.LinkModules()
p.ctx.LinkManagers()

checker := NewFiledChecker(p.ctx)
checker.Check()              // ❌ ResolvedType 为 nil，panic！

// 必须先完成 Pass 2 + Pass 3
p.ctx.BuildSymbolTable()
p.ctx.ResolveReferences()
```

**正确顺序**：
```go
p.ctx.LinkModules()
p.ctx.LinkManagers()
p.ctx.BuildSymbolTable()     // Pass 2
p.ctx.ResolveReferences()    // Pass 3
checker := NewFiledChecker(p.ctx)
checker.Check()              // ✅ ResolvedType 已填充
```

---

## 示例五：新增代码生成器扩展点

如果要为 blueprint_gen 添加新的代码生成器（如生成 TypeScript 类型定义），参考以下模式：

```go
// 1. 创建生成器，接受 *BlueprintContext
type TSGenerator struct {
    ctx *BlueprintContext
}

func NewTSGenerator(ctx *BlueprintContext) *TSGenerator {
    return &TSGenerator{ctx: ctx}
}

// 2. 遍历 Mechanism（唯一有业务字段的层级）
func (g *TSGenerator) Generate(outputDir string) error {
    for _, mech := range g.ctx.Mechanisms {
        for _, field := range mech.GetFields() {
            if field.Type.IsResolved() {
                // 已解析的复杂类型，可以获取 Definition
                def := field.Type.GetResolvedDefinition()
                _ = def.GetKind()  // DefKindStruct, DefKindMessage 等
            } else if field.Type.IsFieldBaseType() {
                // 基础类型，直接转换
                tsType := goTypeToTS(field.Type.Kind.String())
                _ = tsType
            }
        }
    }
    return nil
}

// 3. 在 cmd.go 的 runBlueprintGen 中调用（在 Pass 3 之后）
tsGen := NewTSGenerator(data)
tsGen.Generate(tsOutput)
```

---

## 附录：YAML 解析格式速查

| 语法 | 含义 |
|-----|-----|
| `int32 Gold: 1` | 字段名 Gold，类型 int32，编号 1 |
| `string Name: 2` | 字段名 Name，类型 string，编号 2 |
| `HeroModule HeroData: 1` | 字段引用 MME 对象 HeroModule |
| `map<int32, HeroModule> Heroes: 1` | 普通 map，key=int32，value=HeroModule |
| `xmap<int32, HeroModule> Heroes: 1` | 扩展 map（支持脏标记追踪） |
| `repeated ItemData Items: 1` | 数组，元素类型 ItemData |
| `HeroRarity Rarity: 3` | 枚举字段 |
| `ServiceZoneTypePlay: '1 [Content:"说明"]'` | 枚举值定义（带注释）|
