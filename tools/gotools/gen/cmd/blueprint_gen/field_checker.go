package blueprint_gen

import (
	"fmt"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/mmeobj"
	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

const (
	ObjectFieldNumberMax = 64
)

type FiledChecker struct {
	ctx        *BlueprintContext
	panicError bool
}

func NewFiledChecker(ctx *BlueprintContext) *FiledChecker {
	return &FiledChecker{
		ctx:        ctx,
		panicError: true,
	}
}

func (f *FiledChecker) SetPanicError(panicError bool) {
	f.panicError = panicError
}

func (f *FiledChecker) Check() {
	f.CheckManagerFields()
	f.CheckModuleFields()
	f.CheckEntityFields()
	f.CheckMechanismFields()
}

func (f *FiledChecker) CommonCheck(obj *mmeobj.MMEObject) {
	filteredFields := make([]*types.Field, 0)
	obj.RangeFields(func(field *types.Field) bool {
		if field.Number > ObjectFieldNumberMax {
			fmt.Printf("object %s field %s number is greater than %d\n", obj.Name, field.Name, ObjectFieldNumberMax)
			if f.panicError {
				panic(fmt.Sprintf("object %s field %s number is greater than %d", obj.Name, field.Name, ObjectFieldNumberMax))
			}
			filteredFields = append(filteredFields, field)
		}
		return true
	})
	for _, field := range filteredFields {
		obj.RemoveField(field)
	}
}

func (f *FiledChecker) CheckManagerFields() {
	for _, manager := range f.ctx.Managers {
		f.CommonCheck(manager.MMEObject)

		filteredFields := make([]*types.Field, 0)
		manager.RangeFields(func(field *types.Field) bool {
			switch {
			case field.IsMapField() || field.IsXMapField():
				valueType := field.Type.ValueType
				if !valueType.IsMMEObjectType() {
					fmt.Printf("manager %s map field %s value type is not mme object\n", manager.Name, field.Name)
					if f.panicError {
						panic(fmt.Sprintf("manager %s map field %s value type is not mme object", manager.Name, field.Name))
					}
					filteredFields = append(filteredFields, field)
				}
				resolvedDefinition := valueType.GetResolvedDefinition()
				if resolvedDefinition == nil {
					fmt.Printf("manager %s map field %s value type is not mme object\n", manager.Name, field.Name)
					if f.panicError {
						panic(fmt.Sprintf("manager %s map field %s value type is not mme object", manager.Name, field.Name))
					}
					filteredFields = append(filteredFields, field)
				}

				if resolvedDefinition.GetKind() != blueprint_types.DefKindModule {
					fmt.Printf("manager %s map field %s value type is not manager\n", manager.Name, field.Name)
					if f.panicError {
						panic(fmt.Sprintf("manager %s map field %s value type is not manager", manager.Name, field.Name))
					}
					filteredFields = append(filteredFields, field)
				}
			case field.IsMMEObjectType():
				resolvedDefinition := field.Type.GetResolvedDefinition()
				if resolvedDefinition == nil {
					fmt.Printf("manager %s map field %s value type is not mme object\n", manager.Name, field.Name)
					if f.panicError {
						panic(fmt.Sprintf("manager %s map field %s value type is not mme object", manager.Name, field.Name))
					}
					filteredFields = append(filteredFields, field)
				}
				if resolvedDefinition.GetKind() != blueprint_types.DefKindModule {
					fmt.Printf("manager %s map field %s value type is not module\n", manager.Name, field.Name)
					if f.panicError {
						panic(fmt.Sprintf("manager %s map field %s value type is not module", manager.Name, field.Name))
					}
					filteredFields = append(filteredFields, field)
				}
			default:
				if f.panicError {
					panic(fmt.Sprintf("manager %s field %s type is not supported", manager.Name, field.Name))
				}
				filteredFields = append(filteredFields, field)

			}
			return true
		})
		for _, field := range filteredFields {
			manager.RemoveField(field)
		}
	}
}

func (f *FiledChecker) CheckModuleFields() {
	for _, module := range f.ctx.Modules {

		f.CommonCheck(module.MMEObject)
		filteredFields := make([]*types.Field, 0)
		module.RangeFields(func(field *types.Field) bool {
			if !field.IsMMEObjectType() {
				fmt.Printf("module %s field %s type is not mme object\n", module.Name, field.Name)
				if f.panicError {
					panic(fmt.Sprintf("module %s field %s type is not mme object", module.Name, field.Name))
				}
				filteredFields = append(filteredFields, field)
			}
			resolvedDefinition := field.Type.GetResolvedDefinition()
			if resolvedDefinition == nil {
				fmt.Printf("module %s field %s type is not mme object\n", module.Name, field.Name)
				if f.panicError {
					panic(fmt.Sprintf("module %s field %s type is not mme object", module.Name, field.Name))
				}
				filteredFields = append(filteredFields, field)
			}
			if resolvedDefinition.GetKind() != blueprint_types.DefKindMechanism {
				fmt.Printf("module %s field %s type is not mechanism\n", module.Name, field.Name)
				if f.panicError {
					panic(fmt.Sprintf("module %s field %s type is not mechanism", module.Name, field.Name))
				}
				filteredFields = append(filteredFields, field)
			}
			return true
		})
		for _, field := range filteredFields {
			module.RemoveField(field)
		}
	}
}

func (f *FiledChecker) CheckEntityFields() {
	for _, entity := range f.ctx.Entities {
		f.CommonCheck(entity.MMEObject)
		filteredFields := make([]*types.Field, 0)
		entity.RangeFields(func(field *types.Field) bool {
			if !field.IsMMEObjectType() {
				fmt.Printf("entity %s field %s type is not mme object\n", entity.Name, field.Name)
				if f.panicError {
					panic(fmt.Sprintf("entity %s field %s type is not mme object", entity.Name, field.Name))
				}
				filteredFields = append(filteredFields, field)
			}
			objectType, ok := f.ctx.GetObjectType(field.Type.GetName())
			if !ok {
				fmt.Printf("entity %s field %s type is not mme object\n", entity.Name, field.Name)
				if f.panicError {
					panic(fmt.Sprintf("entity %s field %s type is not mme object", entity.Name, field.Name))
				}
				filteredFields = append(filteredFields, field)
			}
			if !objectType.IsManager() {
				fmt.Printf("entity %s field %s type is not manager\n", entity.Name, field.Name)
				if f.panicError {
					panic(fmt.Sprintf("entity %s field %s type is not manager", entity.Name, field.Name))
				}
				filteredFields = append(filteredFields, field)
			}
			return true
		})
		for _, field := range filteredFields {
			entity.RemoveField(field)
		}
	}
}

func (f *FiledChecker) CheckMechanismFields() {
	for _, mechanism := range f.ctx.Mechanisms {
		f.CommonCheck(mechanism.MMEObject)
		filteredFields := make([]*types.Field, 0)
		mechanism.RangeFields(func(field *types.Field) bool {
			if field.IsMMEObjectType() {
				fmt.Printf("mechanism %s field %s type is not mme object\n", mechanism.Name, field.Name)
				if f.panicError {
					panic(fmt.Sprintf("mechanism %s field %s type is not mme object", mechanism.Name, field.Name))
				}
				filteredFields = append(filteredFields, field)
			}

			return true
		})

		for _, field := range filteredFields {
			mechanism.RemoveField(field)
		}
	}
}
