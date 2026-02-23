package agent_gen

import (
	"fmt"
	"path/filepath"
	"strings"
)

// generateAgentIModels generates the imodels.go file in the agent output directory.
// It contains IBaseAgent and per-type agent interfaces.
func (g *AgentGenerator) generateAgentIModels(agentOutput string) error {
	var sb strings.Builder
	sb.WriteString(generatedHeader)
	sb.WriteString("\npackage mme_agent\n\n")

	needsImodels := len(g.managers) > 0 || len(g.modules) > 0
	if needsImodels {
		sb.WriteString("import (\n")
		sb.WriteString(fmt.Sprintf("\t%q\n", imodelsImportPath))
		sb.WriteString(")\n\n")
	}

	// IBaseAgent
	sb.WriteString("type IBaseAgent interface {\n")
	sb.WriteString("\tGetWrapper() any\n")
	sb.WriteString("}\n\n")

	// I<Manager>Agent for each manager
	for _, manager := range g.managers {
		managerName := manager.Name
		sb.WriteString(fmt.Sprintf("type I%sAgent interface {\n", managerName))
		sb.WriteString("\tIBaseAgent\n")
		sb.WriteString(fmt.Sprintf("\timodels.I%sLogic\n", managerName))
		sb.WriteString("}\n\n")
	}

	// I<Module>Agent for each module
	for _, module := range g.modules {
		moduleName := module.Name
		sb.WriteString(fmt.Sprintf("type I%sAgent interface {\n", moduleName))
		sb.WriteString("\tIBaseAgent\n")

		mechFields := getMechanismFields(module.Fields)
		for _, mf := range mechFields {
			mechTypeName := mf.typeName
			sb.WriteString(fmt.Sprintf("\tGet%sLogic() imodels.I%sLogic\n", mechTypeName, mechTypeName))
		}

		sb.WriteString("}\n\n")
	}

	filePath := filepath.Join(agentOutput, "imodels.go")
	return writeFile(filePath, sb.String())
}

