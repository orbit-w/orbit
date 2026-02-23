package agent_gen

import (
	"fmt"
	"path/filepath"
	"strings"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/mmeobj"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

type managerFieldInfo struct {
	field    *types.Field
	typeName string // Manager type name, e.g. "HeroManager"
}

func getManagerFields(fields []*types.Field) []managerFieldInfo {
	var result []managerFieldInfo
	for _, field := range fields {
		if field.Type.IsMMEObjectType() && field.Type.IsResolved() {
			def := field.Type.GetResolvedDefinition()
			if def != nil && def.GetKind() == types.DefKindManager {
				result = append(result, managerFieldInfo{
					field:    field,
					typeName: field.Type.GetName(),
				})
			}
		}
	}
	return result
}

func (g *AgentGenerator) generateEntityAgent(entity *mmeobj.Entity, agentOutput string) error {
	entityName := entity.Name
	implName := entityName + "Impl"
	managerFields := getManagerFields(entity.Fields)

	var sb strings.Builder
	sb.WriteString(generatedHeader)
	sb.WriteString("\npackage mme_agent\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString(fmt.Sprintf("\tmmeobj %q\n", mmeImportPath))
	sb.WriteString(fmt.Sprintf("\t%q\n", protoMmeImportPath))
	sb.WriteString(fmt.Sprintf("\t%q\n", bsonImportPath))
	sb.WriteString(")\n\n")

	// init() - factory registration
	entityTypeEnum := entityName + "Type"
	sb.WriteString("func init() {\n")
	sb.WriteString(fmt.Sprintf("\tRegisterEntityFactory(mme.EntityType_%s, func() IEntity {\n", entityTypeEnum))
	sb.WriteString(fmt.Sprintf("\t\treturn New%s()\n", entityName))
	sb.WriteString("\t})\n")
	sb.WriteString("}\n\n")

	// Struct
	sb.WriteString(fmt.Sprintf("type %s struct {\n", implName))
	sb.WriteString(fmt.Sprintf("\tentityWrapper *mmeobj.%sWrapper\n", entityName))
	if len(managerFields) > 0 {
		sb.WriteString("\n")
		for _, mf := range managerFields {
			fieldName := firstLower(mf.field.Name) + "Agent"
			sb.WriteString(fmt.Sprintf("\t%s I%sAgent\n", fieldName, mf.typeName))
		}
	}
	sb.WriteString("}\n\n")

	// Constructor
	sb.WriteString(fmt.Sprintf("func New%s() *%s {\n", entityName, implName))
	sb.WriteString(fmt.Sprintf("\treturn &%s{\n", implName))
	sb.WriteString(fmt.Sprintf("\t\tentityWrapper: mmeobj.New%sWrapper(),\n", entityName))
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// GetId
	sb.WriteString(fmt.Sprintf("func (a *%s) GetId() int64 {\n", implName))
	sb.WriteString("\treturn a.entityWrapper.GetXXXId()\n")
	sb.WriteString("}\n\n")

	// GetEntityType
	sb.WriteString(fmt.Sprintf("func (a *%s) GetEntityType() mme.EntityType {\n", implName))
	sb.WriteString("\treturn a.entityWrapper.GetEntityType()\n")
	sb.WriteString("}\n\n")

	// GetEntityWrapper
	sb.WriteString(fmt.Sprintf("func (a *%s) GetEntityWrapper() mmeobj.IEntityWrapper {\n", implName))
	sb.WriteString("\treturn a.entityWrapper\n")
	sb.WriteString("}\n\n")

	// Lazy getter for each manager
	for _, mf := range managerFields {
		fieldName := firstLower(mf.field.Name) + "Agent"
		getterOnWrapper := "Get" + mf.field.Name

		sb.WriteString(fmt.Sprintf("func (a *%s) Get%sAgent() I%sAgent {\n", implName, mf.field.Name, mf.typeName))
		sb.WriteString(fmt.Sprintf("\tif a.%s == nil {\n", fieldName))
		sb.WriteString(fmt.Sprintf("\t\ta.%s = New%sAgent(a.entityWrapper.%s())\n", fieldName, mf.typeName, getterOnWrapper))
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\treturn a.%s\n", fieldName))
		sb.WriteString("}\n\n")
	}

	// OnLoad
	sb.WriteString(fmt.Sprintf("func (a *%s) OnLoad(raw bson.Raw, new bool) error {\n", implName))
	sb.WriteString("\tif err := a.entityWrapper.Load(raw); err != nil {\n")
	sb.WriteString("\t\treturn err\n")
	sb.WriteString("\t}\n\n")
	for _, mf := range managerFields {
		sb.WriteString(fmt.Sprintf("\tif err := CallOnLoad(a.Get%sAgent(), new); err != nil {\n", mf.field.Name))
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n")
	}
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")

	// OnSave
	sb.WriteString(fmt.Sprintf("func (a *%s) OnSave() error {\n", implName))
	for _, mf := range managerFields {
		sb.WriteString(fmt.Sprintf("\tif err := CallOnSave(a.Get%sAgent()); err != nil {\n", mf.field.Name))
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n")
	}
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")

	// OnLogin
	sb.WriteString(fmt.Sprintf("func (a *%s) OnLogin() error {\n", implName))
	for _, mf := range managerFields {
		sb.WriteString(fmt.Sprintf("\tif err := CallOnLogin(a.Get%sAgent()); err != nil {\n", mf.field.Name))
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n")
	}
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")

	// OnLogout
	sb.WriteString(fmt.Sprintf("func (a *%s) OnLogout() error {\n", implName))
	for _, mf := range managerFields {
		sb.WriteString(fmt.Sprintf("\tif err := CallOnLogout(a.Get%sAgent()); err != nil {\n", mf.field.Name))
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n")
	}
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n")

	filePath := filepath.Join(agentOutput, camelToSnake(entityName)+"_impl.go")
	return writeFile(filePath, sb.String())
}
