// Code generated by protoc-gen-go.
// source: golang.conradwood.net/apis/gitbuilder/gitbuilder.proto
// DO NOT EDIT!

/*
Package gitbuilder is a generated protocol buffer package.

It is generated from these files:
	golang.conradwood.net/apis/gitbuilder/gitbuilder.proto

It has these top-level messages:
	BuildRequest
	BuildResponse
*/
package gitbuilder

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "golang.conradwood.net/apis/common"

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

type BuildRequest struct {
	GitURL       string   `protobuf:"bytes,1,opt,name=GitURL" json:"GitURL,omitempty"`
	FetchURLS    []string `protobuf:"bytes,2,rep,name=FetchURLS" json:"FetchURLS,omitempty"`
	CommitID     string   `protobuf:"bytes,3,opt,name=CommitID" json:"CommitID,omitempty"`
	BuildNumber  uint64   `protobuf:"varint,4,opt,name=BuildNumber" json:"BuildNumber,omitempty"`
	RepositoryID uint64   `protobuf:"varint,5,opt,name=RepositoryID" json:"RepositoryID,omitempty"`
	RepoName     string   `protobuf:"bytes,6,opt,name=RepoName" json:"RepoName,omitempty"`
	ArtefactName string   `protobuf:"bytes,7,opt,name=ArtefactName" json:"ArtefactName,omitempty"`
}

func (m *BuildRequest) Reset()                    { *m = BuildRequest{} }
func (m *BuildRequest) String() string            { return proto.CompactTextString(m) }
func (*BuildRequest) ProtoMessage()               {}
func (*BuildRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *BuildRequest) GetGitURL() string {
	if m != nil {
		return m.GitURL
	}
	return ""
}

func (m *BuildRequest) GetFetchURLS() []string {
	if m != nil {
		return m.FetchURLS
	}
	return nil
}

func (m *BuildRequest) GetCommitID() string {
	if m != nil {
		return m.CommitID
	}
	return ""
}

func (m *BuildRequest) GetBuildNumber() uint64 {
	if m != nil {
		return m.BuildNumber
	}
	return 0
}

func (m *BuildRequest) GetRepositoryID() uint64 {
	if m != nil {
		return m.RepositoryID
	}
	return 0
}

func (m *BuildRequest) GetRepoName() string {
	if m != nil {
		return m.RepoName
	}
	return ""
}

func (m *BuildRequest) GetArtefactName() string {
	if m != nil {
		return m.ArtefactName
	}
	return ""
}

type BuildResponse struct {
	Stdout        []byte `protobuf:"bytes,1,opt,name=Stdout,proto3" json:"Stdout,omitempty"`
	Complete      bool   `protobuf:"varint,2,opt,name=Complete" json:"Complete,omitempty"`
	ResultMessage string `protobuf:"bytes,3,opt,name=ResultMessage" json:"ResultMessage,omitempty"`
	Success       bool   `protobuf:"varint,4,opt,name=Success" json:"Success,omitempty"`
}

func (m *BuildResponse) Reset()                    { *m = BuildResponse{} }
func (m *BuildResponse) String() string            { return proto.CompactTextString(m) }
func (*BuildResponse) ProtoMessage()               {}
func (*BuildResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *BuildResponse) GetStdout() []byte {
	if m != nil {
		return m.Stdout
	}
	return nil
}

func (m *BuildResponse) GetComplete() bool {
	if m != nil {
		return m.Complete
	}
	return false
}

func (m *BuildResponse) GetResultMessage() string {
	if m != nil {
		return m.ResultMessage
	}
	return ""
}

func (m *BuildResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func init() {
	proto.RegisterType((*BuildRequest)(nil), "gitbuilder.BuildRequest")
	proto.RegisterType((*BuildResponse)(nil), "gitbuilder.BuildResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for GitBuilder service

type GitBuilderClient interface {
	// build something. Note that this RPC might take several minutes to complete
	Build(ctx context.Context, in *BuildRequest, opts ...grpc.CallOption) (GitBuilder_BuildClient, error)
}

type gitBuilderClient struct {
	cc *grpc.ClientConn
}

func NewGitBuilderClient(cc *grpc.ClientConn) GitBuilderClient {
	return &gitBuilderClient{cc}
}

func (c *gitBuilderClient) Build(ctx context.Context, in *BuildRequest, opts ...grpc.CallOption) (GitBuilder_BuildClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_GitBuilder_serviceDesc.Streams[0], c.cc, "/gitbuilder.GitBuilder/Build", opts...)
	if err != nil {
		return nil, err
	}
	x := &gitBuilderBuildClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GitBuilder_BuildClient interface {
	Recv() (*BuildResponse, error)
	grpc.ClientStream
}

type gitBuilderBuildClient struct {
	grpc.ClientStream
}

func (x *gitBuilderBuildClient) Recv() (*BuildResponse, error) {
	m := new(BuildResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for GitBuilder service

type GitBuilderServer interface {
	// build something. Note that this RPC might take several minutes to complete
	Build(*BuildRequest, GitBuilder_BuildServer) error
}

func RegisterGitBuilderServer(s *grpc.Server, srv GitBuilderServer) {
	s.RegisterService(&_GitBuilder_serviceDesc, srv)
}

func _GitBuilder_Build_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BuildRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GitBuilderServer).Build(m, &gitBuilderBuildServer{stream})
}

type GitBuilder_BuildServer interface {
	Send(*BuildResponse) error
	grpc.ServerStream
}

type gitBuilderBuildServer struct {
	grpc.ServerStream
}

func (x *gitBuilderBuildServer) Send(m *BuildResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _GitBuilder_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitbuilder.GitBuilder",
	HandlerType: (*GitBuilderServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Build",
			Handler:       _GitBuilder_Build_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "golang.conradwood.net/apis/gitbuilder/gitbuilder.proto",
}

func init() {
	proto.RegisterFile("golang.conradwood.net/apis/gitbuilder/gitbuilder.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 344 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x91, 0xcf, 0x4a, 0xf3, 0x40,
	0x14, 0xc5, 0x99, 0xfe, 0xef, 0xfd, 0xda, 0xcd, 0x2c, 0x3e, 0xc6, 0x22, 0x12, 0x8a, 0x8b, 0xac,
	0x52, 0x51, 0x70, 0x29, 0x58, 0x8b, 0xa5, 0x50, 0xbb, 0x98, 0xd2, 0x07, 0x48, 0x93, 0x6b, 0x1c,
	0x48, 0x32, 0x31, 0x73, 0x83, 0xb8, 0xf6, 0x5d, 0x7d, 0x0e, 0xc9, 0x24, 0x6d, 0x53, 0x10, 0x57,
	0xb9, 0xbf, 0x73, 0x73, 0x72, 0x73, 0x38, 0x70, 0x1f, 0xe9, 0xd8, 0x4f, 0x23, 0x2f, 0xd0, 0x69,
	0xee, 0x87, 0x1f, 0x5a, 0x87, 0x5e, 0x8a, 0x34, 0xf3, 0x33, 0x65, 0x66, 0x91, 0xa2, 0x7d, 0xa1,
	0xe2, 0x10, 0xf3, 0xc6, 0xe8, 0x65, 0xb9, 0x26, 0xcd, 0xe1, 0xa4, 0x4c, 0xbc, 0x3f, 0xbe, 0x11,
	0xe8, 0x24, 0xd1, 0x69, 0xfd, 0xa8, 0xbc, 0xd3, 0x6f, 0x06, 0xa3, 0x79, 0xe9, 0x95, 0xf8, 0x5e,
	0xa0, 0x21, 0xfe, 0x1f, 0x7a, 0x4b, 0x45, 0x3b, 0xb9, 0x16, 0xcc, 0x61, 0xee, 0x50, 0xd6, 0xc4,
	0x2f, 0x61, 0xf8, 0x8c, 0x14, 0xbc, 0xed, 0xe4, 0x7a, 0x2b, 0x5a, 0x4e, 0xdb, 0x1d, 0xca, 0x93,
	0xc0, 0x27, 0x30, 0x78, 0xd2, 0x49, 0xa2, 0x68, 0xb5, 0x10, 0x6d, 0xeb, 0x3b, 0x32, 0x77, 0xe0,
	0x9f, 0xbd, 0xb0, 0x29, 0x92, 0x3d, 0xe6, 0xa2, 0xe3, 0x30, 0xb7, 0x23, 0x9b, 0x12, 0x9f, 0xc2,
	0x48, 0x62, 0xa6, 0x8d, 0x22, 0x9d, 0x7f, 0xae, 0x16, 0xa2, 0x6b, 0x5f, 0x39, 0xd3, 0xca, 0x0b,
	0x25, 0x6f, 0xfc, 0x04, 0x45, 0xaf, 0xba, 0x70, 0xe0, 0xd2, 0xff, 0x98, 0x13, 0xbe, 0xfa, 0x01,
	0xd9, 0x7d, 0xdf, 0xee, 0xcf, 0xb4, 0xe9, 0x17, 0x83, 0x71, 0x1d, 0xd4, 0x64, 0x3a, 0x35, 0x58,
	0x26, 0xdd, 0x52, 0xa8, 0x0b, 0xb2, 0x49, 0x47, 0xb2, 0xa6, 0x3a, 0x4b, 0x16, 0x23, 0xa1, 0x68,
	0x39, 0xcc, 0x1d, 0xc8, 0x23, 0xf3, 0x6b, 0x18, 0x4b, 0x34, 0x45, 0x4c, 0x2f, 0x68, 0x8c, 0x1f,
	0x61, 0x1d, 0xf6, 0x5c, 0xe4, 0x02, 0xfa, 0xdb, 0x22, 0x08, 0xd0, 0x18, 0x9b, 0x76, 0x20, 0x0f,
	0x78, 0xbb, 0x06, 0x58, 0x2a, 0x9a, 0x57, 0x65, 0xf1, 0x07, 0xe8, 0xda, 0x91, 0x0b, 0xaf, 0x51,
	0x6a, 0xb3, 0x8e, 0xc9, 0xc5, 0x2f, 0x9b, 0xea, 0xff, 0x6f, 0xd8, 0xdc, 0x81, 0xab, 0x14, 0xa9,
	0xd9, 0x75, 0xd9, 0x73, 0xc3, 0xb1, 0xef, 0xd9, 0x96, 0xef, 0x7e, 0x02, 0x00, 0x00, 0xff, 0xff,
	0xf7, 0xaa, 0x2c, 0x33, 0x5b, 0x02, 0x00, 0x00,
}
