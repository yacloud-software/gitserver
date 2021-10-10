// Code generated by protoc-gen-go.
// source: conradwood.net/apis/proberrepo/proberrepo.proto
// DO NOT EDIT!

/*
Package proberrepo is a generated protocol buffer package.

It is generated from these files:
	conradwood.net/apis/proberrepo/proberrepo.proto

It has these top-level messages:
	PingResponse
*/
package proberrepo

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

func init() {
	proto.RegisterType((*PingResponse)(nil), "proberrepo.PingResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ProberRepoService service

type ProberRepoServiceClient interface {
	// comment: rpc ping
	Ping(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*PingResponse, error)
}

type proberRepoServiceClient struct {
	cc *grpc.ClientConn
}

func NewProberRepoServiceClient(cc *grpc.ClientConn) ProberRepoServiceClient {
	return &proberRepoServiceClient{cc}
}

func (c *proberRepoServiceClient) Ping(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := grpc.Invoke(ctx, "/proberrepo.ProberRepoService/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ProberRepoService service

type ProberRepoServiceServer interface {
	// comment: rpc ping
	Ping(context.Context, *common.Void) (*PingResponse, error)
}

func RegisterProberRepoServiceServer(s *grpc.Server, srv ProberRepoServiceServer) {
	s.RegisterService(&_ProberRepoService_serviceDesc, srv)
}

func _ProberRepoService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(common.Void)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProberRepoServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proberrepo.ProberRepoService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProberRepoServiceServer).Ping(ctx, req.(*common.Void))
	}
	return interceptor(ctx, in, info, handler)
}

var _ProberRepoService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proberrepo.ProberRepoService",
	HandlerType: (*ProberRepoServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _ProberRepoService_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "conradwood.net/apis/proberrepo/proberrepo.proto",
}

func init() { proto.RegisterFile("conradwood.net/apis/proberrepo/proberrepo.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 179 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xd2, 0x4f, 0xce, 0xcf, 0x2b,
	0x4a, 0x4c, 0x29, 0xcf, 0xcf, 0x4f, 0xd1, 0xcb, 0x4b, 0x2d, 0xd1, 0x4f, 0x2c, 0xc8, 0x2c, 0xd6,
	0x2f, 0x28, 0xca, 0x4f, 0x4a, 0x2d, 0x2a, 0x4a, 0x2d, 0xc8, 0x47, 0x62, 0xea, 0x15, 0x14, 0xe5,
	0x97, 0xe4, 0x0b, 0x71, 0x21, 0x44, 0xa4, 0xf4, 0xd2, 0xf3, 0x73, 0x12, 0xf3, 0xd2, 0xf5, 0xb0,
	0x99, 0x91, 0x9c, 0x9f, 0x9b, 0x9b, 0x9f, 0x07, 0xa5, 0x20, 0x7a, 0x95, 0xb4, 0xb8, 0x78, 0x02,
	0x32, 0xf3, 0xd2, 0x83, 0x52, 0x8b, 0x0b, 0xf2, 0xf3, 0x8a, 0x53, 0x85, 0xa4, 0xb8, 0x38, 0x60,
	0x6c, 0x09, 0x46, 0x05, 0x46, 0x0d, 0xce, 0x20, 0x38, 0xdf, 0xc8, 0x99, 0x4b, 0x30, 0x00, 0x6c,
	0x53, 0x50, 0x6a, 0x41, 0x7e, 0x70, 0x6a, 0x51, 0x59, 0x66, 0x72, 0xaa, 0x90, 0x1e, 0x17, 0x0b,
	0xc8, 0x00, 0x21, 0x1e, 0x3d, 0xa8, 0xb9, 0x61, 0xf9, 0x99, 0x29, 0x52, 0x12, 0x7a, 0x48, 0xae,
	0x44, 0xb6, 0xc0, 0x49, 0x81, 0x4b, 0x2e, 0x2f, 0xb5, 0x04, 0xd9, 0x7d, 0x20, 0xb7, 0x21, 0x29,
	0x4f, 0x62, 0x03, 0xbb, 0xcc, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x74, 0x25, 0xfb, 0x21, 0x08,
	0x01, 0x00, 0x00,
}
