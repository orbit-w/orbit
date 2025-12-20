# Ref Extractors 生成器

## 功能说明

此生成器根据 Blueprint 定义中的 NetWall Request 消息，自动生成 `ref_factories.go` 文件。该文件包含从请求消息中提取 `EntityRef` 引用的工具函数。

## 生成内容

生成的 `ref_factories.go` 文件包含：

1. **ExtractRefsFromRequest 类型**：定义了从请求消息中提取 EntityRef 的函数签名
2. **init 函数**：为每个 Request 消息注册提取器函数
3. **RegisterRefExtract 函数**：用于注册自定义的 Ref 提取器
4. **GetExtract 函数**：根据协议 ID 获取对应的 Ref 提取器

## 工作原理

1. 遍历所有 NetWall 定义中的 Request 消息
2. 检测每个 Request 消息中包含 `EntityRef` 类型的字段
3. 为每个 Request 生成对应的提取器函数，自动提取所有 EntityRef 字段
4. 将生成的代码写入 `pkg/proto/pb/ref_factories.go`

## 字段检测

生成器会自动检测以下类型的字段：
- `mme.EntityRef`
- `*mme.EntityRef`
- `EntityRef`
- `*EntityRef`

## 使用示例

生成的代码样例：

```go
func init() {
    RegisterRefExtract(PID_Request_LoginRequest, func(request proto.Message) ([]*mme.EntityRef, error) {
        req := request.(*core.Request_LoginRequest)
        refs := make([]*mme.EntityRef, 0)
        if req.PlayerEntityRef != nil {
            refs = append(refs, req.PlayerEntityRef)
        }
        return refs, nil
    })
}
```

## 集成方式

此生成器已集成到 Blueprint 代码生成流程中，在执行 `blueprintgen` 命令时自动运行。

生成顺序：
1. protocol_ids.pb.go
2. pb_factories.go
3. **ref_factories.go** (此生成器)
4. Go 结构体文件

## 输出位置

默认输出到：`pkg/proto/pb/ref_factories.go`

