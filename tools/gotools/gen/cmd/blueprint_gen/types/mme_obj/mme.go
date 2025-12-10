package mmeobject

import (
	"fmt"
	"strings"
)

// ObjectType MME 对象类型
type ObjectType string

const (
	ObjectTypeUnknown   ObjectType = "Unknown"
	ObjectTypeEntity    ObjectType = "Entity"
	ObjectTypeManager   ObjectType = "Manager"
	ObjectTypeModule    ObjectType = "Module"
	ObjectTypeMechanism ObjectType = "Mechanism"
)

var (
	mmeObjectTypes = []ObjectType{ObjectTypeEntity, ObjectTypeManager, ObjectTypeModule, ObjectTypeMechanism}
)

func (o ObjectType) String() string {
	return string(o)
}

func (o ObjectType) IsEntity() bool {
	return o == ObjectTypeEntity
}

func (o ObjectType) IsManager() bool {
	return o == ObjectTypeManager
}

func (o ObjectType) IsModule() bool {
	return o == ObjectTypeModule
}

func (o ObjectType) IsMechanism() bool {
	return o == ObjectTypeMechanism
}

// IsMMEObjectType 判断类型名称是否是 MME Object 类型
// MME Object 包括：Entity（实体）、Manager（管理器）、Module（模块）、Mechanism（机制）
// 命名模式通常以这些后缀结尾，如：PlayerEntity, HeroManager, HeroModule, HeroMechanism
// 也支持带包名的限定名称，如：mme.HeroManager, core.PlayerEntity
func IsMMEObjectType(typeName string) bool {
	for _, objectType := range mmeObjectTypes {
		if strings.HasSuffix(typeName, objectType.String()) {
			return true
		}
	}
	return false
}

// ParseMMEObjectType 解析类型名称对应的 MME Object 类型
func ParseMMEObjectType(typeName string) (ObjectType, error) {
	for _, objectType := range mmeObjectTypes {
		if strings.HasSuffix(typeName, objectType.String()) {
			return objectType, nil
		}
	}
	return ObjectTypeUnknown, fmt.Errorf("unknown mme object type: %s", typeName)
}
