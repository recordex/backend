// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.2
// source: get_diffpdf.proto

package gen

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetDiffPDFRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetDiffPDFRequest) Reset() {
	*x = GetDiffPDFRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_get_diffpdf_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDiffPDFRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDiffPDFRequest) ProtoMessage() {}

func (x *GetDiffPDFRequest) ProtoReflect() protoreflect.Message {
	mi := &file_get_diffpdf_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDiffPDFRequest.ProtoReflect.Descriptor instead.
func (*GetDiffPDFRequest) Descriptor() ([]byte, []int) {
	return file_get_diffpdf_proto_rawDescGZIP(), []int{0}
}

// GetDiffPDFResponse は何のハッシュ値のファイルと比較したのかをクライアントに知らせるために
// 現時点でブロックチェーンに記録されているそのファイル名の最新バージョンのハッシュ値を返す
type GetDiffPDFResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// newest_file_hash は現時点でブロックチェーンに記録されているそのファイル名の最新バージョンのハッシュ値
	// SHA256 でハッシュ化された 64 文字の 16 進数文字列
	// プリフィックスの 0x を含んで 66 文字
	NewestFileHash string `protobuf:"bytes,1,opt,name=newest_file_hash,json=newestFileHash,proto3" json:"newest_file_hash,omitempty"`
}

func (x *GetDiffPDFResponse) Reset() {
	*x = GetDiffPDFResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_get_diffpdf_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDiffPDFResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDiffPDFResponse) ProtoMessage() {}

func (x *GetDiffPDFResponse) ProtoReflect() protoreflect.Message {
	mi := &file_get_diffpdf_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDiffPDFResponse.ProtoReflect.Descriptor instead.
func (*GetDiffPDFResponse) Descriptor() ([]byte, []int) {
	return file_get_diffpdf_proto_rawDescGZIP(), []int{1}
}

func (x *GetDiffPDFResponse) GetNewestFileHash() string {
	if x != nil {
		return x.NewestFileHash
	}
	return ""
}

var File_get_diffpdf_proto protoreflect.FileDescriptor

var file_get_diffpdf_proto_rawDesc = []byte{
	0x0a, 0x11, 0x67, 0x65, 0x74, 0x5f, 0x64, 0x69, 0x66, 0x66, 0x70, 0x64, 0x66, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x13, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x44, 0x69, 0x66, 0x66, 0x50, 0x44,
	0x46, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x5a, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x44,
	0x69, 0x66, 0x66, 0x50, 0x44, 0x46, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x44,
	0x0a, 0x10, 0x6e, 0x65, 0x77, 0x65, 0x73, 0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x68, 0x61,
	0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1a, 0xfa, 0x42, 0x17, 0x72, 0x15, 0x32,
	0x13, 0x5e, 0x30, 0x78, 0x5b, 0x61, 0x2d, 0x66, 0x41, 0x2d, 0x46, 0x30, 0x2d, 0x39, 0x5d, 0x7b,
	0x36, 0x34, 0x7d, 0x24, 0x52, 0x0e, 0x6e, 0x65, 0x77, 0x65, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65,
	0x48, 0x61, 0x73, 0x68, 0x42, 0x06, 0x5a, 0x04, 0x2f, 0x67, 0x65, 0x6e, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_get_diffpdf_proto_rawDescOnce sync.Once
	file_get_diffpdf_proto_rawDescData = file_get_diffpdf_proto_rawDesc
)

func file_get_diffpdf_proto_rawDescGZIP() []byte {
	file_get_diffpdf_proto_rawDescOnce.Do(func() {
		file_get_diffpdf_proto_rawDescData = protoimpl.X.CompressGZIP(file_get_diffpdf_proto_rawDescData)
	})
	return file_get_diffpdf_proto_rawDescData
}

var file_get_diffpdf_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_get_diffpdf_proto_goTypes = []interface{}{
	(*GetDiffPDFRequest)(nil),  // 0: proto.GetDiffPDFRequest
	(*GetDiffPDFResponse)(nil), // 1: proto.GetDiffPDFResponse
}
var file_get_diffpdf_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_get_diffpdf_proto_init() }
func file_get_diffpdf_proto_init() {
	if File_get_diffpdf_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_get_diffpdf_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDiffPDFRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_get_diffpdf_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDiffPDFResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_get_diffpdf_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_get_diffpdf_proto_goTypes,
		DependencyIndexes: file_get_diffpdf_proto_depIdxs,
		MessageInfos:      file_get_diffpdf_proto_msgTypes,
	}.Build()
	File_get_diffpdf_proto = out.File
	file_get_diffpdf_proto_rawDesc = nil
	file_get_diffpdf_proto_goTypes = nil
	file_get_diffpdf_proto_depIdxs = nil
}
