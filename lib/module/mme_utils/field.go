package mmeutils

import (
	"fmt"

	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
)

type SyncContext int8

const (
	SyncContextServer SyncContext = iota // 服务器同步
	SyncContextClient                    // 客户端同步
)

type IFieldMetaContext interface {
	IsDirty(dirtyBit int64) bool
	MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool
}

// 判断字段是否可以被增量同步
// context: 字段元数据上下文
// dirtyBit: 脏位
// fieldID: 字段ID
// syncContext: 同步上下文
// 返回: 是否可以被增量同步
func FieldCanBeIncrementalSynced(context IFieldMetaContext, dirtyBit int64, fieldID uint8, syncContext SyncContext) bool {
	if !context.IsDirty(dirtyBit) {
		return false
	}

	switch syncContext {
	case SyncContextServer:
		return true
	case SyncContextClient:
		return context.MatchesAll(fieldID, fieldmeta.FieldTypeSync)
	default:
		panic(fmt.Sprintf("invalid sync context: %d", syncContext))
	}
}

// 判断字段是否可以被全量同步
// context: 字段元数据上下文
// fieldID: 字段ID
// syncContext: 同步上下文
// 返回: 是否可以被全量同步
func FieldCanBeFullSynced(context IFieldMetaContext, fieldID uint8, syncContext SyncContext) bool {
	switch syncContext {
	case SyncContextServer:
		return true
	case SyncContextClient:
		return context.MatchesAll(fieldID, fieldmeta.FieldTypeSync)
	default:
		panic(fmt.Sprintf("invalid sync context: %d", syncContext))
	}
}
