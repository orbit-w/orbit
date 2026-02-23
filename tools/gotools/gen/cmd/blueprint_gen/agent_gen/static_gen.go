package agent_gen

import (
	"fmt"
	"path/filepath"
	"strings"
)

// generateStaticFiles generates framework files that don't depend on specific MME types.
func (g *AgentGenerator) generateStaticFiles(agentOutput string) error {
	if err := g.generateEntFactory(agentOutput); err != nil {
		return fmt.Errorf("generate ent_factory: %w", err)
	}
	return nil
}

func (g *AgentGenerator) generateEntFactory(agentOutput string) error {
	var sb strings.Builder
	sb.WriteString(generatedHeader)
	sb.WriteString("\npackage mme_agent\n\n")
	sb.WriteString("import (\n")
	sb.WriteString(fmt.Sprintf("\tmmeobj %q\n", mmeImportPath))
	sb.WriteString(fmt.Sprintf("\t%q\n", protoMmeImportPath))
	sb.WriteString(fmt.Sprintf("\t%q\n", bsonImportPath))
	sb.WriteString(")\n\n")

	sb.WriteString("type IEntity interface {\n")
	sb.WriteString("\tGetId() int64\n")
	sb.WriteString("\tGetEntityType() mme.EntityType\n")
	sb.WriteString("\tGetEntityWrapper() mmeobj.IEntityWrapper\n")
	sb.WriteString("\tOnLoad(raw bson.Raw, new bool) error\n")
	sb.WriteString("\tOnSave() error\n")
	sb.WriteString("\tOnLogin() error\n")
	sb.WriteString("\tOnLogout() error\n")
	sb.WriteString("}\n\n")

	sb.WriteString("type EntityFactory func() IEntity\n\n")
	sb.WriteString("var (\n")
	sb.WriteString("\tmapEntityFactories = make(map[mme.EntityType]EntityFactory)\n")
	sb.WriteString(")\n\n")

	sb.WriteString("func RegisterEntityFactory(entityType mme.EntityType, factory EntityFactory) {\n")
	sb.WriteString("\tmapEntityFactories[entityType] = factory\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func GetEntityFactory(entityType mme.EntityType) EntityFactory {\n")
	sb.WriteString("\tfactory, ok := mapEntityFactories[entityType]\n")
	sb.WriteString("\tif !ok {\n")
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn factory\n")
	sb.WriteString("}\n")

	filePath := filepath.Join(agentOutput, "ent_factory.go")
	return writeFile(filePath, sb.String())
}
