package container

import "gitee.com/orbit-w/orbit/lib/module/xmapwrapper"

type ModuleMapContainer[K comparable, PbValue any, WrapperValue xmapwrapper.Linkable[PbValue], LogicValue any] struct {
	container    xmapwrapper.XMapContainer[K, PbValue, WrapperValue]
	logicFactory func(wrapper WrapperValue) LogicValue
}

func NewModuleMapContainer[K comparable, PbValue any, WrapperValue xmapwrapper.Linkable[PbValue], LogicValue any](container xmapwrapper.XMapContainer[K, PbValue, WrapperValue], logicFactory func(wrapper WrapperValue) LogicValue) *ModuleMapContainer[K, PbValue, WrapperValue, LogicValue] {
	return &ModuleMapContainer[K, PbValue, WrapperValue, LogicValue]{
		container:    container,
		logicFactory: logicFactory,
	}
}

func (c *ModuleMapContainer[K, PbValue, WrapperValue, LogicValue]) Get(key K) (LogicValue, bool) {

	wrapper, exists := c.container.Get(key)
	if !exists {
		var zero LogicValue
		return zero, false
	}
	logic := c.logicFactory(wrapper)
	return logic, true
}

func (c *ModuleMapContainer[K, PbValue, WrapperValue, LogicValue]) Range(fn func(key K, logic LogicValue) bool) {
	c.container.Range(func(key K, wrapper WrapperValue) bool {
		logic := c.logicFactory(wrapper)
		if !fn(key, logic) {
			return false
		}
		return true
	})
}

func (c *ModuleMapContainer[K, PbValue, WrapperValue, LogicValue]) Len() int {
	return c.container.Len()
}
