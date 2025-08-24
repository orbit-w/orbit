package cmd

import gogodesc "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

type Context struct {
	ProtoDir        string
	OutputDir       string
	ProtobufInclude string
	pbfd            *gogodesc.FileDescriptorSet
	mapKeyTypes     []MapKeyTypeInfo
	protoMessages   []ProtoMessage
	MsgIndex        map[string]*gogodesc.DescriptorProto
}

func NewContext(protoDir, outputDir, protobufInclude string) *Context {
	ctx := &Context{
		ProtoDir:        protoDir,
		OutputDir:       outputDir,
		ProtobufInclude: protobufInclude,
		MsgIndex:        make(map[string]*gogodesc.DescriptorProto),
	}
	return ctx
}

func (ctx *Context) SetFileDescriptorSet(pbfd *gogodesc.FileDescriptorSet) {
	ctx.pbfd = pbfd
}

func (ctx *Context) GetFileDescriptorSet() *gogodesc.FileDescriptorSet {
	return ctx.pbfd
}

func (ctx *Context) GetProtoDir() string {
	return ctx.ProtoDir
}

func (ctx *Context) GetOutputDir() string {
	return ctx.OutputDir
}

func (ctx *Context) GetProtobufInclude() string {
	return ctx.ProtobufInclude
}

func (ctx *Context) GetMapKeyTypes() []MapKeyTypeInfo {
	return ctx.mapKeyTypes
}

func (ctx *Context) AddMapKeyType(mapKeyType MapKeyTypeInfo) {
	ctx.mapKeyTypes = append(ctx.mapKeyTypes, mapKeyType)
}

func (ctx *Context) GetProtoMessages() []ProtoMessage {
	return ctx.protoMessages
}

func (ctx *Context) AddProtoMessage(protoMessages ...ProtoMessage) {
	ctx.protoMessages = append(ctx.protoMessages, protoMessages...)
}

func (ctx *Context) Clear() {
	ctx.pbfd = nil
	ctx.mapKeyTypes = nil
	ctx.protoMessages = nil
	ctx.MsgIndex = nil
}

func (ctx *Context) AddMessageToIndex(name string, msg *gogodesc.DescriptorProto) {
	ctx.MsgIndex[name] = msg
}

func (ctx *Context) GetMessageFromIndex(name string) (*gogodesc.DescriptorProto, bool) {
	msg, ok := ctx.MsgIndex[name]
	return msg, ok
}

// Message index accessors
func (ctx *Context) SetMessageIndex(idx map[string]*gogodesc.DescriptorProto) { ctx.MsgIndex = idx }
func (ctx *Context) GetMessageIndex() map[string]*gogodesc.DescriptorProto    { return ctx.MsgIndex }

type StructContext struct {
	protoDir        string
	outputDir       string
	protobufInclude string
}

func NewStructContext(protoDir, outputDir, protobufInclude string) *StructContext {
	return &StructContext{protoDir: protoDir, outputDir: outputDir, protobufInclude: protobufInclude}
}
