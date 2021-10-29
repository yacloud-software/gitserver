// Code generated by protoc-gen-go.
// source: golang.conradwood.net/apis/gitdiffsync/gitdiffsync.proto
// DO NOT EDIT!

/*
Package gitdiffsync is a generated protocol buffer package.

It is generated from these files:
	golang.conradwood.net/apis/gitdiffsync/gitdiffsync.proto

It has these top-level messages:
	PingResponse
	AddSyncRepoRequest
	SyncRepo
	SyncDir
	PatchApply
	LastStatus
*/
package gitdiffsync

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import common "golang.conradwood.net/apis/common"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SyncType int32

const (
	SyncType_Git    SyncType = 0
	SyncType_Gerrit SyncType = 1
)

var SyncType_name = map[int32]string{
	0: "Git",
	1: "Gerrit",
}
var SyncType_value = map[string]int32{
	"Git":    0,
	"Gerrit": 1,
}

func (x SyncType) String() string {
	return proto.EnumName(SyncType_name, int32(x))
}
func (SyncType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// comment: message pingresponse
type PingResponse struct {
	// comment: field pingresponse.response
	Response string `protobuf:"bytes,1,opt,name=Response" json:"Response,omitempty"`
}

func (m *PingResponse) Reset()                    { *m = PingResponse{} }
func (m *PingResponse) String() string            { return proto.CompactTextString(m) }
func (*PingResponse) ProtoMessage()               {}
func (*PingResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *PingResponse) GetResponse() string {
	if m != nil {
		return m.Response
	}
	return ""
}

type AddSyncRepoRequest struct {
	SourceRepository string   `protobuf:"bytes,1,opt,name=SourceRepository" json:"SourceRepository,omitempty"`
	TargetRepository string   `protobuf:"bytes,2,opt,name=TargetRepository" json:"TargetRepository,omitempty"`
	SyncType         SyncType `protobuf:"varint,3,opt,name=SyncType,enum=gitdiffsync.SyncType" json:"SyncType,omitempty"`
	Directories      []string `protobuf:"bytes,4,rep,name=Directories" json:"Directories,omitempty"`
	ForceCopy        bool     `protobuf:"varint,5,opt,name=ForceCopy" json:"ForceCopy,omitempty"`
}

func (m *AddSyncRepoRequest) Reset()                    { *m = AddSyncRepoRequest{} }
func (m *AddSyncRepoRequest) String() string            { return proto.CompactTextString(m) }
func (*AddSyncRepoRequest) ProtoMessage()               {}
func (*AddSyncRepoRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *AddSyncRepoRequest) GetSourceRepository() string {
	if m != nil {
		return m.SourceRepository
	}
	return ""
}

func (m *AddSyncRepoRequest) GetTargetRepository() string {
	if m != nil {
		return m.TargetRepository
	}
	return ""
}

func (m *AddSyncRepoRequest) GetSyncType() SyncType {
	if m != nil {
		return m.SyncType
	}
	return SyncType_Git
}

func (m *AddSyncRepoRequest) GetDirectories() []string {
	if m != nil {
		return m.Directories
	}
	return nil
}

func (m *AddSyncRepoRequest) GetForceCopy() bool {
	if m != nil {
		return m.ForceCopy
	}
	return false
}

// configure a repo to sync
type SyncRepo struct {
	ID               uint64   `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	RepositoryID     uint64   `protobuf:"varint,2,opt,name=RepositoryID" json:"RepositoryID,omitempty"`
	UserID           string   `protobuf:"bytes,3,opt,name=UserID" json:"UserID,omitempty"`
	TargetURL        string   `protobuf:"bytes,4,opt,name=TargetURL" json:"TargetURL,omitempty"`
	SyncType         SyncType `protobuf:"varint,5,opt,name=SyncType,enum=gitdiffsync.SyncType" json:"SyncType,omitempty"`
	LastCommitSynced string   `protobuf:"bytes,6,opt,name=LastCommitSynced" json:"LastCommitSynced,omitempty"`
}

func (m *SyncRepo) Reset()                    { *m = SyncRepo{} }
func (m *SyncRepo) String() string            { return proto.CompactTextString(m) }
func (*SyncRepo) ProtoMessage()               {}
func (*SyncRepo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *SyncRepo) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *SyncRepo) GetRepositoryID() uint64 {
	if m != nil {
		return m.RepositoryID
	}
	return 0
}

func (m *SyncRepo) GetUserID() string {
	if m != nil {
		return m.UserID
	}
	return ""
}

func (m *SyncRepo) GetTargetURL() string {
	if m != nil {
		return m.TargetURL
	}
	return ""
}

func (m *SyncRepo) GetSyncType() SyncType {
	if m != nil {
		return m.SyncType
	}
	return SyncType_Git
}

func (m *SyncRepo) GetLastCommitSynced() string {
	if m != nil {
		return m.LastCommitSynced
	}
	return ""
}

type SyncDir struct {
	ID               uint64    `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	SyncRepo         *SyncRepo `protobuf:"bytes,2,opt,name=SyncRepo" json:"SyncRepo,omitempty"`
	Dir              string    `protobuf:"bytes,3,opt,name=Dir" json:"Dir,omitempty"`
	LastCommitSynced string    `protobuf:"bytes,4,opt,name=LastCommitSynced" json:"LastCommitSynced,omitempty"`
	ForceCopy        bool      `protobuf:"varint,5,opt,name=ForceCopy" json:"ForceCopy,omitempty"`
}

func (m *SyncDir) Reset()                    { *m = SyncDir{} }
func (m *SyncDir) String() string            { return proto.CompactTextString(m) }
func (*SyncDir) ProtoMessage()               {}
func (*SyncDir) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *SyncDir) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *SyncDir) GetSyncRepo() *SyncRepo {
	if m != nil {
		return m.SyncRepo
	}
	return nil
}

func (m *SyncDir) GetDir() string {
	if m != nil {
		return m.Dir
	}
	return ""
}

func (m *SyncDir) GetLastCommitSynced() string {
	if m != nil {
		return m.LastCommitSynced
	}
	return ""
}

func (m *SyncDir) GetForceCopy() bool {
	if m != nil {
		return m.ForceCopy
	}
	return false
}

type PatchApply struct {
	ID           uint64    `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	SyncRepo     *SyncRepo `protobuf:"bytes,2,opt,name=SyncRepo" json:"SyncRepo,omitempty"`
	Timestamp    uint32    `protobuf:"varint,3,opt,name=Timestamp" json:"Timestamp,omitempty"`
	SourceCommit string    `protobuf:"bytes,4,opt,name=SourceCommit" json:"SourceCommit,omitempty"`
	Success      bool      `protobuf:"varint,5,opt,name=Success" json:"Success,omitempty"`
	Message      string    `protobuf:"bytes,6,opt,name=Message" json:"Message,omitempty"`
}

func (m *PatchApply) Reset()                    { *m = PatchApply{} }
func (m *PatchApply) String() string            { return proto.CompactTextString(m) }
func (*PatchApply) ProtoMessage()               {}
func (*PatchApply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PatchApply) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *PatchApply) GetSyncRepo() *SyncRepo {
	if m != nil {
		return m.SyncRepo
	}
	return nil
}

func (m *PatchApply) GetTimestamp() uint32 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *PatchApply) GetSourceCommit() string {
	if m != nil {
		return m.SourceCommit
	}
	return ""
}

func (m *PatchApply) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *PatchApply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type LastStatus struct {
	ID        uint64    `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	SyncRepo  *SyncRepo `protobuf:"bytes,2,opt,name=SyncRepo" json:"SyncRepo,omitempty"`
	Timestamp uint32    `protobuf:"varint,3,opt,name=Timestamp" json:"Timestamp,omitempty"`
	Success   bool      `protobuf:"varint,4,opt,name=Success" json:"Success,omitempty"`
	Message   string    `protobuf:"bytes,5,opt,name=Message" json:"Message,omitempty"`
}

func (m *LastStatus) Reset()                    { *m = LastStatus{} }
func (m *LastStatus) String() string            { return proto.CompactTextString(m) }
func (*LastStatus) ProtoMessage()               {}
func (*LastStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *LastStatus) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *LastStatus) GetSyncRepo() *SyncRepo {
	if m != nil {
		return m.SyncRepo
	}
	return nil
}

func (m *LastStatus) GetTimestamp() uint32 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *LastStatus) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *LastStatus) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*PingResponse)(nil), "gitdiffsync.PingResponse")
	proto.RegisterType((*AddSyncRepoRequest)(nil), "gitdiffsync.AddSyncRepoRequest")
	proto.RegisterType((*SyncRepo)(nil), "gitdiffsync.SyncRepo")
	proto.RegisterType((*SyncDir)(nil), "gitdiffsync.SyncDir")
	proto.RegisterType((*PatchApply)(nil), "gitdiffsync.PatchApply")
	proto.RegisterType((*LastStatus)(nil), "gitdiffsync.LastStatus")
	proto.RegisterEnum("gitdiffsync.SyncType", SyncType_name, SyncType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for GitDiffSync service

type GitDiffSyncClient interface {
	AddSyncRepo(ctx context.Context, in *AddSyncRepoRequest, opts ...grpc.CallOption) (*common.Void, error)
	Trigger(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*common.Void, error)
}

type gitDiffSyncClient struct {
	cc *grpc.ClientConn
}

func NewGitDiffSyncClient(cc *grpc.ClientConn) GitDiffSyncClient {
	return &gitDiffSyncClient{cc}
}

func (c *gitDiffSyncClient) AddSyncRepo(ctx context.Context, in *AddSyncRepoRequest, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/gitdiffsync.GitDiffSync/AddSyncRepo", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitDiffSyncClient) Trigger(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/gitdiffsync.GitDiffSync/Trigger", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for GitDiffSync service

type GitDiffSyncServer interface {
	AddSyncRepo(context.Context, *AddSyncRepoRequest) (*common.Void, error)
	Trigger(context.Context, *common.Void) (*common.Void, error)
}

func RegisterGitDiffSyncServer(s *grpc.Server, srv GitDiffSyncServer) {
	s.RegisterService(&_GitDiffSync_serviceDesc, srv)
}

func _GitDiffSync_AddSyncRepo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddSyncRepoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitDiffSyncServer).AddSyncRepo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitdiffsync.GitDiffSync/AddSyncRepo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitDiffSyncServer).AddSyncRepo(ctx, req.(*AddSyncRepoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitDiffSync_Trigger_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(common.Void)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitDiffSyncServer).Trigger(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitdiffsync.GitDiffSync/Trigger",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitDiffSyncServer).Trigger(ctx, req.(*common.Void))
	}
	return interceptor(ctx, in, info, handler)
}

var _GitDiffSync_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitdiffsync.GitDiffSync",
	HandlerType: (*GitDiffSyncServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddSyncRepo",
			Handler:    _GitDiffSync_AddSyncRepo_Handler,
		},
		{
			MethodName: "Trigger",
			Handler:    _GitDiffSync_Trigger_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "golang.conradwood.net/apis/gitdiffsync/gitdiffsync.proto",
}

func init() {
	proto.RegisterFile("golang.conradwood.net/apis/gitdiffsync/gitdiffsync.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 581 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xb4, 0x94, 0xdf, 0x8a, 0xd3, 0x4e,
	0x14, 0xc7, 0x7f, 0x69, 0xd3, 0x7f, 0x27, 0xfd, 0x2d, 0x65, 0x40, 0x49, 0x8b, 0xd0, 0x6e, 0x40,
	0x28, 0xbd, 0xc8, 0xe2, 0x2a, 0xe2, 0x85, 0x37, 0xbb, 0x1b, 0x2c, 0x85, 0x0a, 0xcb, 0xb4, 0xeb,
	0x7d, 0x4c, 0xa6, 0x71, 0x60, 0x93, 0x89, 0x33, 0x13, 0x24, 0xb7, 0x7d, 0x0c, 0x2f, 0xbd, 0xf4,
	0x35, 0xf4, 0x0d, 0x7c, 0x00, 0x7d, 0x14, 0x99, 0x24, 0xdd, 0x24, 0xdb, 0xae, 0xf6, 0xc6, 0xab,
	0x9e, 0xf3, 0x3d, 0x27, 0x3d, 0xdf, 0xcf, 0xcc, 0x70, 0xe0, 0x55, 0xc0, 0x6e, 0xdd, 0x28, 0xb0,
	0x3d, 0x16, 0x71, 0xd7, 0xff, 0xc4, 0x98, 0x6f, 0x47, 0x44, 0x9e, 0xb9, 0x31, 0x15, 0x67, 0x01,
	0x95, 0x3e, 0xdd, 0x6c, 0x44, 0x1a, 0x79, 0xd5, 0xd8, 0x8e, 0x39, 0x93, 0x0c, 0x19, 0x15, 0x69,
	0x64, 0xff, 0xe1, 0x6f, 0x3c, 0x16, 0x86, 0x2c, 0x2a, 0x7e, 0xf2, 0x8f, 0xad, 0x19, 0xf4, 0xaf,
	0x69, 0x14, 0x60, 0x22, 0x62, 0x16, 0x09, 0x82, 0x46, 0xd0, 0xdd, 0xc5, 0xa6, 0x36, 0xd1, 0xa6,
	0x3d, 0x7c, 0x97, 0x5b, 0x3f, 0x35, 0x40, 0x17, 0xbe, 0xbf, 0x4a, 0x23, 0x0f, 0x93, 0x98, 0x61,
	0xf2, 0x31, 0x21, 0x42, 0xa2, 0x19, 0x0c, 0x56, 0x2c, 0xe1, 0x1e, 0x51, 0xa2, 0xa0, 0x92, 0xf1,
	0xb4, 0xf8, 0x74, 0x4f, 0x57, 0xbd, 0x6b, 0x97, 0x07, 0x44, 0x56, 0x7a, 0x1b, 0x79, 0xef, 0x7d,
	0x1d, 0x3d, 0x83, 0xae, 0x1a, 0xb5, 0x4e, 0x63, 0x62, 0x36, 0x27, 0xda, 0xf4, 0xe4, 0xfc, 0x91,
	0x5d, 0xa5, 0xdf, 0x15, 0xf1, 0x5d, 0x1b, 0x9a, 0x80, 0xe1, 0x50, 0x4e, 0x3c, 0xc9, 0x38, 0x25,
	0xc2, 0xd4, 0x27, 0xcd, 0x69, 0x0f, 0x57, 0x25, 0xf4, 0x04, 0x7a, 0x6f, 0x18, 0xf7, 0xc8, 0x15,
	0x8b, 0x53, 0xb3, 0x35, 0xd1, 0xa6, 0x5d, 0x5c, 0x0a, 0xd6, 0x0f, 0x2d, 0x9f, 0xa9, 0x5c, 0xa0,
	0x13, 0x68, 0x2c, 0x9c, 0x8c, 0x44, 0xc7, 0x8d, 0x85, 0x83, 0x2c, 0xe8, 0x97, 0xee, 0x16, 0x4e,
	0xe6, 0x5b, 0xc7, 0x35, 0x0d, 0x3d, 0x86, 0xf6, 0x8d, 0x20, 0x7c, 0xe1, 0x64, 0x8e, 0x7b, 0xb8,
	0xc8, 0xd4, 0xd8, 0x9c, 0xef, 0x06, 0x2f, 0x4d, 0x3d, 0x2b, 0x95, 0x42, 0x8d, 0xb4, 0x75, 0x1c,
	0xe9, 0x0c, 0x06, 0x4b, 0x57, 0xc8, 0x2b, 0x16, 0x86, 0x54, 0x2a, 0x95, 0xf8, 0x66, 0x3b, 0x3f,
	0xc8, 0xfb, 0xba, 0xf5, 0x4d, 0x83, 0x8e, 0x0a, 0x1d, 0xca, 0xf7, 0xa0, 0x96, 0x25, 0x70, 0x06,
	0x64, 0x1c, 0x18, 0xad, 0x8a, 0x97, 0xc3, 0xcf, 0xdb, 0x61, 0x3b, 0xa1, 0x91, 0x7c, 0xf9, 0xe2,
	0xcb, 0x76, 0x68, 0xa8, 0x22, 0x27, 0x31, 0xb3, 0xa9, 0x8f, 0xcb, 0x23, 0x1b, 0x40, 0xd3, 0xa1,
	0xbc, 0x60, 0x57, 0xe1, 0x41, 0x9f, 0xfa, 0x61, 0x9f, 0x7f, 0xb9, 0x9b, 0x5f, 0x1a, 0xc0, 0xb5,
	0x2b, 0xbd, 0x0f, 0x17, 0x71, 0x7c, 0x9b, 0xfe, 0x63, 0x10, 0x75, 0x5f, 0x34, 0x24, 0x42, 0xba,
	0x61, 0x9c, 0xe1, 0xfc, 0x8f, 0x4b, 0x41, 0xbd, 0x84, 0xfc, 0x65, 0xe7, 0xf6, 0x0b, 0xa0, 0x9a,
	0x86, 0x4c, 0xe8, 0xac, 0x12, 0xcf, 0x23, 0x42, 0x14, 0x28, 0xbb, 0x54, 0x55, 0xde, 0x12, 0x21,
	0xdc, 0x80, 0x14, 0x37, 0xb6, 0x4b, 0xad, 0xef, 0x1a, 0x80, 0x3a, 0x95, 0x95, 0x74, 0x65, 0x22,
	0xf6, 0x10, 0xd7, 0xc7, 0x22, 0x9e, 0x3e, 0x88, 0xf8, 0x75, 0x3b, 0xd4, 0x25, 0x4f, 0xc8, 0xd1,
	0xa8, 0x15, 0x0c, 0xfd, 0x41, 0x8c, 0x56, 0x0d, 0x63, 0x36, 0x2e, 0x9f, 0x33, 0xea, 0x40, 0x73,
	0x4e, 0xe5, 0xe0, 0x3f, 0x04, 0xd0, 0x9e, 0x13, 0xce, 0xa9, 0x1c, 0x68, 0xe7, 0x1c, 0x8c, 0x39,
	0x95, 0x0e, 0xdd, 0x6c, 0x54, 0x1f, 0x7a, 0x0d, 0x46, 0x65, 0xad, 0xa0, 0x71, 0x0d, 0x6a, 0x7f,
	0xe1, 0x8c, 0xfa, 0x76, 0xb1, 0xc2, 0xde, 0x31, 0xea, 0xa3, 0xa7, 0xd0, 0x59, 0x73, 0x1a, 0x04,
	0x84, 0xa3, 0x5a, 0xa1, 0xde, 0x76, 0x79, 0x0a, 0xe3, 0x88, 0xc8, 0xea, 0x5e, 0x54, 0x3b, 0xb1,
	0x3a, 0xea, 0x7d, 0x3b, 0x5b, 0x89, 0xcf, 0x7f, 0x07, 0x00, 0x00, 0xff, 0xff, 0x1f, 0x45, 0x1b,
	0x04, 0x8b, 0x05, 0x00, 0x00,
}
