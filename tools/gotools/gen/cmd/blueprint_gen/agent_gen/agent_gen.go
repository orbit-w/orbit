package agent_gen

import (
	"fmt"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/mmeobj"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// AgentGenerator 代理层代码生成器
// 生成 mme_agent/ 和 mme_logic/ 目录下的所有代理层代码
type AgentGenerator struct {
	entities    []*mmeobj.Entity
	managers    []*mmeobj.Manager
	modules     []*mmeobj.Module
	mechanisms  []*mmeobj.Mechanism
	symbolTable *types.SymbolTable
}

// NewAgentGenerator 创建代理层代码生成器
func NewAgentGenerator(
	entities []*mmeobj.Entity,
	managers []*mmeobj.Manager,
	modules []*mmeobj.Module,
	mechanisms []*mmeobj.Mechanism,
	symbolTable *types.SymbolTable,
) *AgentGenerator {
	return &AgentGenerator{
		entities:    entities,
		managers:    managers,
		modules:     modules,
		mechanisms:  mechanisms,
		symbolTable: symbolTable,
	}
}

// Generate 生成所有代理层代码
// agentOutput: mme_agent 输出目录
func (g *AgentGenerator) Generate(agentOutput string) error {
	if err := g.generateStaticFiles(agentOutput); err != nil {
		return fmt.Errorf("generate static files: %w", err)
	}

	if err := g.generateAgentIModels(agentOutput); err != nil {
		return fmt.Errorf("generate agent imodels: %w", err)
	}

	for _, entity := range g.entities {
		if err := g.generateEntityAgent(entity, agentOutput); err != nil {
			return fmt.Errorf("generate entity agent %s: %w", entity.Name, err)
		}
	}

	for _, manager := range g.managers {
		if err := g.generateManagerAgent(manager, agentOutput); err != nil {
			return fmt.Errorf("generate manager agent %s: %w", manager.Name, err)
		}
	}

	for _, module := range g.modules {
		if err := g.generateModuleAgent(module, agentOutput); err != nil {
			return fmt.Errorf("generate module agent %s: %w", module.Name, err)
		}
	}

	return nil
}
