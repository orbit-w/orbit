package blueprint_gen

import (
	"fmt"
)

// Parser YAML 解析器
type Parser struct {
	*BaseParser
}

// NewParser 创建新的解析器
func NewParser(blueprintDir string) *Parser {
	return &Parser{
		BaseParser: NewBaseParser(blueprintDir),
	}
}

// Parse 解析所有 YAML 文件
func (p *Parser) Parse() error {
	// 解析 headfile.yaml
	if err := p.parseHeadFile(); err != nil {
		return fmt.Errorf("failed to parse headfile: %w", err)
	}

	// 解析 MME 文件
	if err := p.parseMMEFiles(); err != nil {
		return fmt.Errorf("failed to parse MME files: %w", err)
	}

	// 解析 NetWall 文件
	if err := p.parseNetWallFiles(); err != nil {
		return fmt.Errorf("failed to parse NetWall files: %w", err)
	}

	// 检查 Entity 字段类型是否符合要求
	p.ctx.CheckEntityFields()

	return nil
}

// parseHeadFile 解析 headfile.yaml
func (p *Parser) parseHeadFile() error {
	return nil
}
