package blueprint_gen

import types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"

type MMEObjectBase interface {
	GetFields() []*types.Field
	HasMapField() bool
	HasXMapField() bool
}
