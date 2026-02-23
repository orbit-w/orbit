package agent_gen

import (
	"fmt"
	"path/filepath"
	"strings"
)

// generateLogicInterfaces generates the logic interface files in mme_logic/imodels/.
// These are regenerated each time (not user-editable).
func (g *AgentGenerator) generateLogicInterfaces(logicOutput string) error {
	imodelsDir := filepath.Join(logicOutput, "imodels")

	// base.go - IBaseLogic interface
	if err := g.generateBaseLogicInterface(imodelsDir); err != nil {
		return fmt.Errorf("generate base logic interface: %w", err)
	}

	// Per-manager logic interfaces
	for _, manager := range g.managers {
		if err := g.generateManagerLogicInterface(manager.Name, imodelsDir); err != nil {
			return fmt.Errorf("generate manager logic interface %s: %w", manager.Name, err)
		}
	}

	// Per-mechanism logic interfaces
	for _, mechanism := range g.mechanisms {
		if err := g.generateMechanismLogicInterface(mechanism.Name, imodelsDir); err != nil {
			return fmt.Errorf("generate mechanism logic interface %s: %w", mechanism.Name, err)
		}
	}

	return nil
}

func (g *AgentGenerator) generateBaseLogicInterface(imodelsDir string) error {
	var sb strings.Builder
	sb.WriteString(generatedHeader)
	sb.WriteString("\npackage imodels\n\n")
	sb.WriteString("type IBaseLogic interface {\n")
	sb.WriteString("\tGetWrapper() any\n")
	sb.WriteString("}\n")

	filePath := filepath.Join(imodelsDir, "base.go")
	return writeFile(filePath, sb.String())
}

func (g *AgentGenerator) generateManagerLogicInterface(managerName string, imodelsDir string) error {
	var sb strings.Builder
	sb.WriteString(generatedHeader)
	sb.WriteString("\npackage imodels\n\n")
	sb.WriteString(fmt.Sprintf("type I%sLogic interface {\n", managerName))
	sb.WriteString("}\n")

	fileName := "i_" + camelToSnake(managerName) + "_logic.go"
	filePath := filepath.Join(imodelsDir, fileName)
	return writeFile(filePath, sb.String())
}

func (g *AgentGenerator) generateMechanismLogicInterface(mechName string, imodelsDir string) error {
	var sb strings.Builder
	sb.WriteString(generatedHeader)
	sb.WriteString("\npackage imodels\n\n")
	sb.WriteString(fmt.Sprintf("type I%sLogic interface {\n", mechName))
	sb.WriteString("\tIBaseLogic\n")
	sb.WriteString("}\n")

	fileName := "i_" + camelToSnake(mechName) + "_logic.go"
	filePath := filepath.Join(imodelsDir, fileName)
	return writeFile(filePath, sb.String())
}

// generateLogicScaffolds generates implementation scaffold files ONLY if they don't exist.
// These files are user-editable and never overwritten.
func (g *AgentGenerator) generateLogicScaffolds(logicOutput string) error {
	// Manager logic scaffolds
	managersDir := filepath.Join(logicOutput, "managers")
	for _, manager := range g.managers {
		if err := g.generateManagerLogicScaffold(manager.Name, managersDir); err != nil {
			return fmt.Errorf("generate manager logic scaffold %s: %w", manager.Name, err)
		}
	}

	// Mechanism logic scaffolds
	mechanismsDir := filepath.Join(logicOutput, "mechanisms")
	for _, mechanism := range g.mechanisms {
		if err := g.generateMechanismLogicScaffold(mechanism.Name, mechanismsDir); err != nil {
			return fmt.Errorf("generate mechanism logic scaffold %s: %w", mechanism.Name, err)
		}
	}

	return nil
}

func (g *AgentGenerator) generateManagerLogicScaffold(managerName string, managersDir string) error {
	fileName := camelToSnake(managerName) + "_logic_impl.go"
	filePath := filepath.Join(managersDir, fileName)
	if fileExists(filePath) {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("package managers\n\n")
	sb.WriteString("import (\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", mmeImportPath))
	sb.WriteString(fmt.Sprintf("\t%q\n", imodelsImportPath))
	sb.WriteString(")\n\n")

	implName := managerName + "LogicImpl"
	sb.WriteString(fmt.Sprintf("type %s struct {\n", implName))
	sb.WriteString(fmt.Sprintf("\twrapper *mme.%sWrapper\n", managerName))
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("func New%sLogic(wrapper *mme.%sWrapper) imodels.I%sLogic {\n",
		managerName, managerName, managerName))
	sb.WriteString(fmt.Sprintf("\treturn &%s{\n", implName))
	sb.WriteString("\t\twrapper: wrapper,\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return writeFile(filePath, sb.String())
}

func (g *AgentGenerator) generateMechanismLogicScaffold(mechName string, mechanismsDir string) error {
	fileName := camelToSnake(mechName) + "_logic_impl.go"
	filePath := filepath.Join(mechanismsDir, fileName)
	if fileExists(filePath) {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("package mechanisms\n\n")
	sb.WriteString("import (\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", mmeImportPath))
	sb.WriteString(fmt.Sprintf("\t%q\n", imodelsImportPath))
	sb.WriteString(")\n\n")

	implName := mechName + "LogicImpl"
	sb.WriteString(fmt.Sprintf("type %s struct {\n", implName))
	sb.WriteString(fmt.Sprintf("\twrapper *mme.%sWrapper\n", mechName))
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("func New%sLogic(wrapper *mme.%sWrapper) imodels.I%sLogic {\n",
		mechName, mechName, mechName))
	sb.WriteString(fmt.Sprintf("\treturn &%s{\n", implName))
	sb.WriteString("\t\twrapper: wrapper,\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("func (m *%s) GetWrapper() any {\n", implName))
	sb.WriteString("\treturn m.wrapper\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("func (m *%s) OnLoad(new bool) error { return nil }\n\n", implName))
	sb.WriteString(fmt.Sprintf("func (m *%s) OnSave() error { return nil }\n\n", implName))
	sb.WriteString(fmt.Sprintf("func (m *%s) OnLogin() error { return nil }\n\n", implName))
	sb.WriteString(fmt.Sprintf("func (m *%s) OnLogout() error { return nil }\n", implName))

	return writeFile(filePath, sb.String())
}
