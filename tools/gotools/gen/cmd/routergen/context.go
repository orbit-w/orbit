package routergen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gogo/protobuf/proto"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

type Context struct {
	outputDir       string
	protobufInclude string
	descFile        string
	protoFiles      []string
	fds             descriptor.FileDescriptorSet
	fdsParsed       bool // 标记是否已经解析过fds
	mode            Mode
	genProtoIDs     bool
	genProtoCode    bool
}

func NewContext(protoFiles []string, outputDir, protobufInclude string) *Context {
	return &Context{
		outputDir:       outputDir,
		protobufInclude: protobufInclude,
		protoFiles:      protoFiles,
		mode:            ModeNormal,
		genProtoIDs:     true,
		genProtoCode:    true,
	}
}

func (c *Context) SetDescFile(descFile string) {
	c.descFile = descFile
}

func (c *Context) UnmarshalFds(data []byte) error {
	var fds descriptor.FileDescriptorSet
	if err := proto.Unmarshal(data, &fds); err != nil {
		return fmt.Errorf("unmarshaling descriptor: %w", err)
	}
	c.fds = fds
	return nil
}

func (c *Context) GetFds() descriptor.FileDescriptorSet {
	return c.fds
}

func (c *Context) GetOutputDir() string {
	return c.outputDir
}

func (c *Context) GetProtoFiles() []string {
	return c.protoFiles
}

// ParseProtoFiles 执行protoc命令并解析fds，只执行一次
func (c *Context) ParseProtoFiles() ([]*descriptor.FileDescriptorProto, error) {
	if c.fdsParsed {
		// 如果已经解析过，直接返回缓存的结果
		return c.fds.File, nil
	}

	if len(c.protoFiles) == 0 {
		return nil, fmt.Errorf("no proto files provided")
	}

	descFile := "temp.desc"                   // Use a fixed temp file or make unique if needed
	protoDir := filepath.Dir(c.protoFiles[0]) // Assume same dir, or handle multiple

	var cmdArgs []string
	cmdArgs = append(cmdArgs,
		"--proto_path=.",
		"--proto_path="+protoDir,
		"--proto_path=vendor/github.com/asynkron/protoactor-go/actor",
		"--proto_path=$GOPATH/pkg/mod",
		"--descriptor_set_out="+descFile,
		"--include_imports", // Include all dependencies
	)
	cmdArgs = append(cmdArgs, c.protoFiles...)

	cmd := exec.Command("protoc", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("protoc failed: %w\nOutput: %s", err, output)
	}
	defer os.Remove(descFile)

	data, err := os.ReadFile(descFile)
	if err != nil {
		return nil, fmt.Errorf("reading descriptor: %w", err)
	}

	if err := proto.Unmarshal(data, &c.fds); err != nil {
		return nil, fmt.Errorf("unmarshaling descriptor: %w", err)
	}

	if len(c.fds.File) == 0 {
		return nil, fmt.Errorf("no files in descriptor set")
	}

	c.fdsParsed = true
	c.descFile = descFile
	return c.fds.File, nil
}

// SetMode 设置运行模式
func (c *Context) SetMode(mode Mode) {
	c.mode = mode
}

// GetMode 获取运行模式
func (c *Context) GetMode() Mode {
	return c.mode
}

// SetGenProtoIDs 设置是否生成协议ID
func (c *Context) SetGenProtoIDs(gen bool) {
	c.genProtoIDs = gen
}

// GetGenProtoIDs 获取是否生成协议ID
func (c *Context) GetGenProtoIDs() bool {
	return c.genProtoIDs
}

// SetGenProtoCode 设置是否生成协议代码
func (c *Context) SetGenProtoCode(gen bool) {
	c.genProtoCode = gen
}

// GetGenProtoCode 获取是否生成协议代码
func (c *Context) GetGenProtoCode() bool {
	return c.genProtoCode
}
