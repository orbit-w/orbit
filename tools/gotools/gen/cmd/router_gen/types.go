package router_gen

// EntityRefInfo EntityRef 字段信息
type EntityRefInfo struct {
	FieldName    string // 字段名，如 "PlayerEntityRef"
	EntityName   string // 实体名，如 "Player"（从 PlayerEntityRef 提取）
	ParamName    string // 参数名，如 "playerEntity"（用于 Handle 方法参数）
	WrapperType  string // 包装器类型，如 "*mme.PlayerEntityWrapper"
	WrapperTypeWithAlias string // 带别名的包装器类型，如 "*mmeobj.PlayerEntityWrapper"
}

// RequestInfo Request 信息
type RequestInfo struct {
	RequestName  string          // Request 名称，如 "LoginRequest"
	PackageName  string          // 包名（小写），如 "core"
	PackageNameUpper string      // 包名（首字母大写），如 "Core"
	EntityRefs   []*EntityRefInfo // EntityRef 字段列表（按出现顺序）
	HasRsp       bool            // 是否有响应
}

// ControllerInfo Controller 信息
type ControllerInfo struct {
	PackageName  string // Controller 包名，如 "controllerv2"
	VarName      string // Controller 变量名，如 "GControllerV2"
	TypeName     string // Controller 类型名，如 "Controller"
}

// RouterGenContext 代码生成上下文
type RouterGenContext struct {
	Requests       []*RequestInfo
	Controller     *ControllerInfo
	OutputPath     string // Router 输出文件路径
	ControllerPath string // Controller 文件路径（可选）
}

