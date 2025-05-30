// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.3
// source: app/proto/example.proto

//消息所属的包，需要和文件名相同

package pb_core

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// ------发送墙，包含的消息可以由客户端发送，由服务端回复rsp
type Request struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Request) Reset() {
	*x = Request{}
	mi := &file_app_proto_example_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_example_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_app_proto_example_proto_rawDescGZIP(), []int{0}
}

// -------通知墙，包含的消息只能由服务器发送给客户端
type Notify struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Notify) Reset() {
	*x = Notify{}
	mi := &file_app_proto_example_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Notify) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Notify) ProtoMessage() {}

func (x *Notify) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_example_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Notify.ProtoReflect.Descriptor instead.
func (*Notify) Descriptor() ([]byte, []int) {
	return file_app_proto_example_proto_rawDescGZIP(), []int{1}
}

// --------在墙外定义的是单纯的数据结构，无法单独发送
type Book struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Content       string                 `protobuf:"bytes,1,opt,name=Content,proto3" json:"Content,omitempty"` //这行注释会被胶水代码读取
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Book) Reset() {
	*x = Book{}
	mi := &file_app_proto_example_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Book) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Book) ProtoMessage() {}

func (x *Book) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_example_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Book.ProtoReflect.Descriptor instead.
func (*Book) Descriptor() ([]byte, []int) {
	return file_app_proto_example_proto_rawDescGZIP(), []int{2}
}

func (x *Book) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

// 通用成功
type OK struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OK) Reset() {
	*x = OK{}
	mi := &file_app_proto_example_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OK) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OK) ProtoMessage() {}

func (x *OK) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_example_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OK.ProtoReflect.Descriptor instead.
func (*OK) Descriptor() ([]byte, []int) {
	return file_app_proto_example_proto_rawDescGZIP(), []int{3}
}

// 通用失败
type Fail struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Reason        string                 `protobuf:"bytes,1,opt,name=Reason,proto3" json:"Reason,omitempty"` //Reason建议命名：error_墙名_消息名_Reason
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Fail) Reset() {
	*x = Fail{}
	mi := &file_app_proto_example_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Fail) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fail) ProtoMessage() {}

func (x *Fail) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_example_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fail.ProtoReflect.Descriptor instead.
func (*Fail) Descriptor() ([]byte, []int) {
	return file_app_proto_example_proto_rawDescGZIP(), []int{4}
}

func (x *Fail) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

// 只有直接放在消息前的注释会被胶水代码读取
// 名字可以随便取，同一个包内不能重名
type Request_SearchBook struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Query         string                 `protobuf:"bytes,1,opt,name=Query,proto3" json:"Query,omitempty"` //这行注释会被胶水代码读取
	PageNumber    int32                  `protobuf:"varint,2,opt,name=PageNumber,proto3" json:"PageNumber,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Request_SearchBook) Reset() {
	*x = Request_SearchBook{}
	mi := &file_app_proto_example_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Request_SearchBook) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request_SearchBook) ProtoMessage() {}

func (x *Request_SearchBook) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_example_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request_SearchBook.ProtoReflect.Descriptor instead.
func (*Request_SearchBook) Descriptor() ([]byte, []int) {
	return file_app_proto_example_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Request_SearchBook) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

func (x *Request_SearchBook) GetPageNumber() int32 {
	if x != nil {
		return x.PageNumber
	}
	return 0
}

// 心跳
type Request_HeartBeat struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Request_HeartBeat) Reset() {
	*x = Request_HeartBeat{}
	mi := &file_app_proto_example_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Request_HeartBeat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request_HeartBeat) ProtoMessage() {}

func (x *Request_HeartBeat) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_example_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request_HeartBeat.ProtoReflect.Descriptor instead.
func (*Request_HeartBeat) Descriptor() ([]byte, []int) {
	return file_app_proto_example_proto_rawDescGZIP(), []int{0, 1}
}

// 该请求的回复消息，名字必须为Rsp，
// 如果没有，则默认回复为通用成功OK
type Request_SearchBook_Rsp struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Result        *Book                  `protobuf:"bytes,1,opt,name=Result,proto3" json:"Result,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Request_SearchBook_Rsp) Reset() {
	*x = Request_SearchBook_Rsp{}
	mi := &file_app_proto_example_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Request_SearchBook_Rsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request_SearchBook_Rsp) ProtoMessage() {}

func (x *Request_SearchBook_Rsp) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_example_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request_SearchBook_Rsp.ProtoReflect.Descriptor instead.
func (*Request_SearchBook_Rsp) Descriptor() ([]byte, []int) {
	return file_app_proto_example_proto_rawDescGZIP(), []int{0, 0, 0}
}

func (x *Request_SearchBook_Rsp) GetResult() *Book {
	if x != nil {
		return x.Result
	}
	return nil
}

type Notify_BeAttacked struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	CurHp         int32                  `protobuf:"varint,1,opt,name=CurHp,proto3" json:"CurHp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Notify_BeAttacked) Reset() {
	*x = Notify_BeAttacked{}
	mi := &file_app_proto_example_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Notify_BeAttacked) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Notify_BeAttacked) ProtoMessage() {}

func (x *Notify_BeAttacked) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_example_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Notify_BeAttacked.ProtoReflect.Descriptor instead.
func (*Notify_BeAttacked) Descriptor() ([]byte, []int) {
	return file_app_proto_example_proto_rawDescGZIP(), []int{1, 0}
}

func (x *Notify_BeAttacked) GetCurHp() int32 {
	if x != nil {
		return x.CurHp
	}
	return 0
}

var File_app_proto_example_proto protoreflect.FileDescriptor

var file_app_proto_example_proto_rawDesc = string([]byte{
	0x0a, 0x17, 0x61, 0x70, 0x70, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x43, 0x6f, 0x72, 0x65, 0x22,
	0x85, 0x01, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x6d, 0x0a, 0x0a, 0x53,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x42, 0x6f, 0x6f, 0x6b, 0x12, 0x14, 0x0a, 0x05, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12,
	0x1e, 0x0a, 0x0a, 0x50, 0x61, 0x67, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0a, 0x50, 0x61, 0x67, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x1a,
	0x29, 0x0a, 0x03, 0x52, 0x73, 0x70, 0x12, 0x22, 0x0a, 0x06, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x43, 0x6f, 0x72, 0x65, 0x2e, 0x42, 0x6f,
	0x6f, 0x6b, 0x52, 0x06, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x1a, 0x0b, 0x0a, 0x09, 0x48, 0x65,
	0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x22, 0x2c, 0x0a, 0x06, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x79, 0x1a, 0x22, 0x0a, 0x0a, 0x42, 0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x65, 0x64, 0x12,
	0x14, 0x0a, 0x05, 0x43, 0x75, 0x72, 0x48, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05,
	0x43, 0x75, 0x72, 0x48, 0x70, 0x22, 0x20, 0x0a, 0x04, 0x42, 0x6f, 0x6f, 0x6b, 0x12, 0x18, 0x0a,
	0x07, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x04, 0x0a, 0x02, 0x4f, 0x4b, 0x22, 0x1e, 0x0a,
	0x04, 0x46, 0x61, 0x69, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x42, 0x0c, 0x5a,
	0x0a, 0x70, 0x62, 0x2f, 0x70, 0x62, 0x5f, 0x63, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
})

var (
	file_app_proto_example_proto_rawDescOnce sync.Once
	file_app_proto_example_proto_rawDescData []byte
)

func file_app_proto_example_proto_rawDescGZIP() []byte {
	file_app_proto_example_proto_rawDescOnce.Do(func() {
		file_app_proto_example_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_proto_example_proto_rawDesc), len(file_app_proto_example_proto_rawDesc)))
	})
	return file_app_proto_example_proto_rawDescData
}

var file_app_proto_example_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_app_proto_example_proto_goTypes = []any{
	(*Request)(nil),                // 0: Core.Request
	(*Notify)(nil),                 // 1: Core.Notify
	(*Book)(nil),                   // 2: Core.Book
	(*OK)(nil),                     // 3: Core.OK
	(*Fail)(nil),                   // 4: Core.Fail
	(*Request_SearchBook)(nil),     // 5: Core.Request.SearchBook
	(*Request_HeartBeat)(nil),      // 6: Core.Request.HeartBeat
	(*Request_SearchBook_Rsp)(nil), // 7: Core.Request.SearchBook.Rsp
	(*Notify_BeAttacked)(nil),      // 8: Core.Notify.BeAttacked
}
var file_app_proto_example_proto_depIdxs = []int32{
	2, // 0: Core.Request.SearchBook.Rsp.Result:type_name -> Core.Book
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_app_proto_example_proto_init() }
func file_app_proto_example_proto_init() {
	if File_app_proto_example_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_proto_example_proto_rawDesc), len(file_app_proto_example_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_proto_example_proto_goTypes,
		DependencyIndexes: file_app_proto_example_proto_depIdxs,
		MessageInfos:      file_app_proto_example_proto_msgTypes,
	}.Build()
	File_app_proto_example_proto = out.File
	file_app_proto_example_proto_goTypes = nil
	file_app_proto_example_proto_depIdxs = nil
}
