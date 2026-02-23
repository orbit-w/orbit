package agent_gen

import (
	"fmt"
	"path/filepath"
	"strings"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/mmeobj"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

type mechanismFieldInfo struct {
	field    *types.Field
	typeName string // Mechanism type name, e.g. "HeroMechanism"
}

func getMechanismFields(fields []*types.Field) []mechanismFieldInfo {
	var result []mechanismFieldInfo
	for _, field := range fields {
		if field.Type.IsMMEObjectType() && field.Type.IsResolved() {
			def := field.Type.GetResolvedDefinition()
			if def != nil && def.GetKind() == types.DefKindMechanism {
				result = append(result, mechanismFieldInfo{
					field:    field,
					typeName: field.Type.GetName(),
				})
			}
		}
	}
	return result
}

func (g *AgentGenerator) generateModuleAgent(module *mmeobj.Module, agentOutput string) error {
	moduleName := module.Name
	implName := moduleName + "AgentImpl"
	mechFields := getMechanismFields(module.Fields)

	var sb strings.Builder
	sb.WriteString(generatedHeader)
	sb.WriteString("\npackage mme_agent\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", mmeImportPath))
	if len(mechFields) > 0 {
		sb.WriteString(fmt.Sprintf("\t%q\n", imodelsImportPath))
		sb.WriteString(fmt.Sprintf("\t%q\n", mechanismsLogicImportPath))
	}
	sb.WriteString(")\n\n")

	// Struct
	sb.WriteString(fmt.Sprintf("type %s struct {\n", implName))
	sb.WriteString(fmt.Sprintf("\twrapper *mme.%sWrapper\n", moduleName))
	if len(mechFields) > 0 {
		sb.WriteString("\n")
		for _, mf := range mechFields {
			fieldName := firstLower(mf.field.Name) + "Logic"
			sb.WriteString(fmt.Sprintf("\t%s imodels.I%sLogic\n", fieldName, mf.typeName))
		}
	}
	sb.WriteString("}\n\n")

	// Constructor
	sb.WriteString(fmt.Sprintf("func New%sAgent(wrapper *mme.%sWrapper) I%sAgent {\n",
		moduleName, moduleName, moduleName))
	sb.WriteString(fmt.Sprintf("\treturn &%s{\n", implName))
	sb.WriteString("\t\twrapper: wrapper,\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// GetWrapper
	sb.WriteString(fmt.Sprintf("func (m *%s) GetWrapper() any {\n", implName))
	sb.WriteString("\treturn m.wrapper\n")
	sb.WriteString("}\n\n")

	// Lifecycle methods
	writeModuleLifecycle(&sb, implName, mechFields, "OnLoad", true)
	writeModuleLifecycle(&sb, implName, mechFields, "OnSave", false)
	writeModuleLifecycle(&sb, implName, mechFields, "OnLogin", false)
	writeModuleLifecycle(&sb, implName, mechFields, "OnLogout", false)

	// Mechanism lazy getters
	for _, mf := range mechFields {
		fieldName := firstLower(mf.field.Name) + "Logic"
		wrapperGetter := "Get" + mf.field.Name
		mechTypeName := mf.typeName

		sb.WriteString(fmt.Sprintf("func (m *%s) Get%sLogic() imodels.I%sLogic {\n",
			implName, mechTypeName, mechTypeName))
		sb.WriteString(fmt.Sprintf("\tif m.%s == nil {\n", fieldName))
		sb.WriteString(fmt.Sprintf("\t\tm.%s = mechanisms.New%sLogic(m.wrapper.%s())\n",
			fieldName, mechTypeName, wrapperGetter))
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\treturn m.%s\n", fieldName))
		sb.WriteString("}\n\n")
	}

	filePath := filepath.Join(agentOutput, camelToSnake(moduleName)+"_agent_impl.go")
	return writeFile(filePath, sb.String())
}

func writeModuleLifecycle(sb *strings.Builder, implName string, mechFields []mechanismFieldInfo, method string, hasNewParam bool) {
	if hasNewParam {
		sb.WriteString(fmt.Sprintf("func (m *%s) %s(new bool) error {\n", implName, method))
	} else {
		sb.WriteString(fmt.Sprintf("func (m *%s) %s() error {\n", implName, method))
	}

	for _, mf := range mechFields {
		mechTypeName := mf.typeName
		getter := fmt.Sprintf("Get%sLogic", mechTypeName)
		if hasNewParam {
			sb.WriteString(fmt.Sprintf("\tif err := Call%s(m.%s(), new); err != nil {\n", method, getter))
		} else {
			sb.WriteString(fmt.Sprintf("\tif err := Call%s(m.%s()); err != nil {\n", method, getter))
		}
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n")
	}

	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
}
