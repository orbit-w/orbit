package imodels

// IMechanismLogic Mechanism Logic 基础接口
type IBaseLogic interface {
	// GetWrapper 获取对应的 Mechanism Wrapper
	GetWrapper() any

	// OnLoad 加载逻辑
	OnLoad(new bool) error
	// OnSave 保存逻辑
	OnSave() error
	// OnLogin 登录回调
	OnLogin() error
	// OnLogout 登出回调
	OnLogout() error
}
