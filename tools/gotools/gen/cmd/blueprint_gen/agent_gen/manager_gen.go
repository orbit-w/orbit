package agent_gen

import (
	"fmt"
	"path/filepath"
	"strings"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/mmeobj"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// moduleFieldKind distinguishes map-container vs direct module fields
type moduleFieldKind int

const (
	moduleFieldMap    moduleFieldKind = iota // xmap or map field
	moduleFieldDirect                        // direct MME object reference
)

type moduleFieldInfo struct {
	field      *types.Field
	kind       moduleFieldKind
	moduleName string // Module type name, e.g. "HeroModule"
	keyType    string // key type for map fields, e.g. "int64"
}

func getModuleFields(fields []*types.Field) []moduleFieldInfo {
	var result []moduleFieldInfo
	for _, field := range fields {
		switch {
		case (field.Type.IsXMapField() || field.Type.IsMapField()):
			vt := field.Type.GetValueType()
			if vt != nil && vt.IsMMEObjectType() && vt.IsResolved() {
				def := vt.GetResolvedDefinition()
				if def != nil && def.GetKind() == types.DefKindModule {
					result = append(result, moduleFieldInfo{
						field:      field,
						kind:       moduleFieldMap,
						moduleName: vt.GetName(),
						keyType:    field.Type.KeyKind().String(),
					})
				}
			}
		case field.Type.IsMMEObjectType() && field.Type.IsResolved():
			def := field.Type.GetResolvedDefinition()
			if def != nil && def.GetKind() == types.DefKindModule {
				result = append(result, moduleFieldInfo{
					field:      field,
					kind:       moduleFieldDirect,
					moduleName: field.Type.GetName(),
				})
			}
		}
	}
	return result
}

func (g *AgentGenerator) generateManagerAgent(manager *mmeobj.Manager, agentOutput string) error {
	managerName := manager.Name
	implName := managerName + "AgentImpl"
	moduleFields := getModuleFields(manager.Fields)

	var sb strings.Builder
	sb.WriteString(generatedHeader)
	sb.WriteString("\npackage mme_agent\n\n")

	// Imports
	needsContainer := false
	for _, mf := range moduleFields {
		if mf.kind == moduleFieldMap {
			needsContainer = true
			break
		}
	}
	sb.WriteString("import (\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", mmeImportPath))
	sb.WriteString(fmt.Sprintf("\t%q\n", imodelsImportPath))
	if needsContainer {
		sb.WriteString(fmt.Sprintf("\t%q\n", containerImportPath))
	}
	sb.WriteString(fmt.Sprintf("\t%q\n", managersLogicImportPath))
	sb.WriteString(")\n\n")

	// Struct
	sb.WriteString(fmt.Sprintf("type %s struct {\n", implName))
	sb.WriteString(fmt.Sprintf("\timodels.I%sLogic\n", managerName))
	sb.WriteString(fmt.Sprintf("\twrapper *mme.%sWrapper\n", managerName))
	if len(moduleFields) > 0 {
		sb.WriteString("\n")
		for _, mf := range moduleFields {
			switch mf.kind {
			case moduleFieldMap:
				containerField := firstLower(mf.field.Name) + "Container"
				sb.WriteString(fmt.Sprintf("\t%s *container.ModuleMapContainer[%s, *mme.%s, *mme.%sWrapper, I%sAgent]\n",
					containerField, mf.keyType, mf.moduleName, mf.moduleName, mf.moduleName))
			case moduleFieldDirect:
				agentField := firstLower(mf.field.Name) + "Agent"
				sb.WriteString(fmt.Sprintf("\t%s I%sAgent\n", agentField, mf.moduleName))
			}
		}
	}
	sb.WriteString("}\n\n")

	// Constructor
	sb.WriteString(fmt.Sprintf("func New%sAgent(wrapper *mme.%sWrapper) I%sAgent {\n",
		managerName, managerName, managerName))
	sb.WriteString(fmt.Sprintf("\tins := &%s{\n", implName))
	sb.WriteString(fmt.Sprintf("\t\twrapper:       wrapper,\n"))
	sb.WriteString(fmt.Sprintf("\t\tI%sLogic: managers.New%sLogic(wrapper),\n", managerName, managerName))
	for _, mf := range moduleFields {
		if mf.kind == moduleFieldMap {
			containerField := firstLower(mf.field.Name) + "Container"
			wrapperGetter := "Get" + mf.field.Name
			sb.WriteString(fmt.Sprintf("\t\t%s: container.NewModuleMapContainer(\n", containerField))
			sb.WriteString(fmt.Sprintf("\t\t\twrapper.%s(),\n", wrapperGetter))
			sb.WriteString(fmt.Sprintf("\t\t\tNew%sAgent,\n", mf.moduleName))
			sb.WriteString("\t\t),\n")
		}
	}
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn ins\n")
	sb.WriteString("}\n\n")

	// GetWrapper
	sb.WriteString(fmt.Sprintf("func (m *%s) GetWrapper() any {\n", implName))
	sb.WriteString("\treturn m.wrapper\n")
	sb.WriteString("}\n\n")

	// Lifecycle methods
	g.writeManagerLifecycle(&sb, implName, moduleFields, "OnLoad", true)
	g.writeManagerLifecycle(&sb, implName, moduleFields, "OnSave", false)
	g.writeManagerLifecycle(&sb, implName, moduleFields, "OnLogin", false)
	g.writeManagerLifecycle(&sb, implName, moduleFields, "OnLogout", false)

	// Map access methods
	for _, mf := range moduleFields {
		if mf.kind == moduleFieldMap {
			containerField := firstLower(mf.field.Name) + "Container"
			sb.WriteString(fmt.Sprintf("func (m *%s) %s_GetModule(key %s) (I%sAgent, bool) {\n",
				implName, mf.field.Name, mf.keyType, mf.moduleName))
			sb.WriteString(fmt.Sprintf("\tlogic, exists := m.%s.Get(key)\n", containerField))
			sb.WriteString("\treturn logic, exists\n")
			sb.WriteString("}\n\n")
		}
	}

	// Direct module lazy getters
	for _, mf := range moduleFields {
		if mf.kind == moduleFieldDirect {
			agentField := firstLower(mf.field.Name) + "Agent"
			wrapperGetter := "Get" + mf.field.Name

			sb.WriteString(fmt.Sprintf("func (m *%s) Get%sAgent() I%sAgent {\n",
				implName, mf.field.Name, mf.moduleName))
			sb.WriteString(fmt.Sprintf("\tif m.%s == nil {\n", agentField))
			sb.WriteString(fmt.Sprintf("\t\twrapper := m.wrapper.%s()\n", wrapperGetter))
			sb.WriteString("\t\tif wrapper == nil {\n")
			sb.WriteString("\t\t\treturn nil\n")
			sb.WriteString("\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\tm.%s = New%sAgent(wrapper)\n", agentField, mf.moduleName))
			sb.WriteString("\t}\n")
			sb.WriteString(fmt.Sprintf("\treturn m.%s\n", agentField))
			sb.WriteString("}\n\n")
		}
	}

	filePath := filepath.Join(agentOutput, camelToSnake(managerName)+"_agent_impl.go")
	return writeFile(filePath, sb.String())
}

func (g *AgentGenerator) writeManagerLifecycle(sb *strings.Builder, implName string, moduleFields []moduleFieldInfo, method string, hasNewParam bool) {
	if hasNewParam {
		sb.WriteString(fmt.Sprintf("func (m *%s) %s(new bool) error {\n", implName, method))
	} else {
		sb.WriteString(fmt.Sprintf("func (m *%s) %s() error {\n", implName, method))
	}

	// Declare err once if we have any map containers
	hasMapField := false
	for _, mf := range moduleFields {
		if mf.kind == moduleFieldMap {
			hasMapField = true
			break
		}
	}
	if hasMapField {
		sb.WriteString("\tvar err error\n")
	}

	// Map containers: Range lifecycle
	for _, mf := range moduleFields {
		if mf.kind == moduleFieldMap {
			containerField := firstLower(mf.field.Name) + "Container"
			sb.WriteString(fmt.Sprintf("\tm.%s.Range(func(key %s, agent I%sAgent) bool {\n",
				containerField, mf.keyType, mf.moduleName))
			if hasNewParam {
				sb.WriteString(fmt.Sprintf("\t\tif err = Call%s(agent, new); err != nil {\n", method))
			} else {
				sb.WriteString(fmt.Sprintf("\t\tif err = Call%s(agent); err != nil {\n", method))
			}
			sb.WriteString("\t\t\treturn false\n")
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t\treturn true\n")
			sb.WriteString("\t})\n")
			sb.WriteString("\tif err != nil {\n")
			sb.WriteString("\t\treturn err\n")
			sb.WriteString("\t}\n\n")
		}
	}

	// Direct modules: CallOnXxx with lazy getter
	for _, mf := range moduleFields {
		if mf.kind == moduleFieldDirect {
			getter := "Get" + mf.field.Name + "Agent"
			if hasNewParam {
				sb.WriteString(fmt.Sprintf("\tif err := Call%s(m.%s(), new); err != nil {\n", method, getter))
			} else {
				sb.WriteString(fmt.Sprintf("\tif err := Call%s(m.%s()); err != nil {\n", method, getter))
			}
			sb.WriteString("\t\treturn err\n")
			sb.WriteString("\t}\n")
		}
	}

	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
}
