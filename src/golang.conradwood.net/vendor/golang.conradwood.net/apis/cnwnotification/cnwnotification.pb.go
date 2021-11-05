// Code generated by protoc-gen-go.
// source: golang.conradwood.net/apis/cnwnotification/cnwnotification.proto
// DO NOT EDIT!

/*
Package cnwnotification is a generated protocol buffer package.

It is generated from these files:
	golang.conradwood.net/apis/cnwnotification/cnwnotification.proto

It has these top-level messages:
	ViaSMSRequest
	ConfigRequest
	ConfigResponse
	Notification
	DisplayRequest
	SoundRequest
	SuppressRequest
	Connection
	ConnectionList
	WifiInfoRequest
	ButtonPressedRequest
	Info
*/
package cnwnotification

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

// send an SMS via some devices' sim card
type ViaSMSRequest struct {
	Number  string `protobuf:"bytes,1,opt,name=Number" json:"Number,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=Message" json:"Message,omitempty"`
}

func (m *ViaSMSRequest) Reset()                    { *m = ViaSMSRequest{} }
func (m *ViaSMSRequest) String() string            { return proto.CompactTextString(m) }
func (*ViaSMSRequest) ProtoMessage()               {}
func (*ViaSMSRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ViaSMSRequest) GetNumber() string {
	if m != nil {
		return m.Number
	}
	return ""
}

func (m *ViaSMSRequest) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type ConfigRequest struct {
	Secret string `protobuf:"bytes,1,opt,name=Secret" json:"Secret,omitempty"`
}

func (m *ConfigRequest) Reset()                    { *m = ConfigRequest{} }
func (m *ConfigRequest) String() string            { return proto.CompactTextString(m) }
func (*ConfigRequest) ProtoMessage()               {}
func (*ConfigRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ConfigRequest) GetSecret() string {
	if m != nil {
		return m.Secret
	}
	return ""
}

type ConfigResponse struct {
	Notifications []*Notification `protobuf:"bytes,1,rep,name=Notifications" json:"Notifications,omitempty"`
	Foo           string          `protobuf:"bytes,2,opt,name=Foo" json:"Foo,omitempty"`
	Bar           string          `protobuf:"bytes,3,opt,name=Bar" json:"Bar,omitempty"`
}

func (m *ConfigResponse) Reset()                    { *m = ConfigResponse{} }
func (m *ConfigResponse) String() string            { return proto.CompactTextString(m) }
func (*ConfigResponse) ProtoMessage()               {}
func (*ConfigResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ConfigResponse) GetNotifications() []*Notification {
	if m != nil {
		return m.Notifications
	}
	return nil
}

func (m *ConfigResponse) GetFoo() string {
	if m != nil {
		return m.Foo
	}
	return ""
}

func (m *ConfigResponse) GetBar() string {
	if m != nil {
		return m.Bar
	}
	return ""
}

type Notification struct {
	Title string `protobuf:"bytes,1,opt,name=Title" json:"Title,omitempty"`
	Text  string `protobuf:"bytes,2,opt,name=Text" json:"Text,omitempty"`
	URL   string `protobuf:"bytes,3,opt,name=URL" json:"URL,omitempty"`
	Sound string `protobuf:"bytes,4,opt,name=Sound" json:"Sound,omitempty"`
	Image string `protobuf:"bytes,5,opt,name=Image" json:"Image,omitempty"`
}

func (m *Notification) Reset()                    { *m = Notification{} }
func (m *Notification) String() string            { return proto.CompactTextString(m) }
func (*Notification) ProtoMessage()               {}
func (*Notification) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Notification) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *Notification) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func (m *Notification) GetURL() string {
	if m != nil {
		return m.URL
	}
	return ""
}

func (m *Notification) GetSound() string {
	if m != nil {
		return m.Sound
	}
	return ""
}

func (m *Notification) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

type DisplayRequest struct {
	Secret     string `protobuf:"bytes,1,opt,name=Secret" json:"Secret,omitempty"`
	Display    string `protobuf:"bytes,2,opt,name=Display" json:"Display,omitempty"`
	Background string `protobuf:"bytes,3,opt,name=Background" json:"Background,omitempty"`
	TextColour string `protobuf:"bytes,4,opt,name=TextColour" json:"TextColour,omitempty"`
}

func (m *DisplayRequest) Reset()                    { *m = DisplayRequest{} }
func (m *DisplayRequest) String() string            { return proto.CompactTextString(m) }
func (*DisplayRequest) ProtoMessage()               {}
func (*DisplayRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *DisplayRequest) GetSecret() string {
	if m != nil {
		return m.Secret
	}
	return ""
}

func (m *DisplayRequest) GetDisplay() string {
	if m != nil {
		return m.Display
	}
	return ""
}

func (m *DisplayRequest) GetBackground() string {
	if m != nil {
		return m.Background
	}
	return ""
}

func (m *DisplayRequest) GetTextColour() string {
	if m != nil {
		return m.TextColour
	}
	return ""
}

type SoundRequest struct {
	Secret        string `protobuf:"bytes,1,opt,name=Secret" json:"Secret,omitempty"`
	URL           string `protobuf:"bytes,2,opt,name=URL" json:"URL,omitempty"`
	Origin        string `protobuf:"bytes,3,opt,name=Origin" json:"Origin,omitempty"`
	SkipQuietZone bool   `protobuf:"varint,4,opt,name=SkipQuietZone" json:"SkipQuietZone,omitempty"`
}

func (m *SoundRequest) Reset()                    { *m = SoundRequest{} }
func (m *SoundRequest) String() string            { return proto.CompactTextString(m) }
func (*SoundRequest) ProtoMessage()               {}
func (*SoundRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *SoundRequest) GetSecret() string {
	if m != nil {
		return m.Secret
	}
	return ""
}

func (m *SoundRequest) GetURL() string {
	if m != nil {
		return m.URL
	}
	return ""
}

func (m *SoundRequest) GetOrigin() string {
	if m != nil {
		return m.Origin
	}
	return ""
}

func (m *SoundRequest) GetSkipQuietZone() bool {
	if m != nil {
		return m.SkipQuietZone
	}
	return false
}

type SuppressRequest struct {
	Secret    string `protobuf:"bytes,1,opt,name=Secret" json:"Secret,omitempty"`
	Timestamp uint32 `protobuf:"varint,2,opt,name=Timestamp" json:"Timestamp,omitempty"`
}

func (m *SuppressRequest) Reset()                    { *m = SuppressRequest{} }
func (m *SuppressRequest) String() string            { return proto.CompactTextString(m) }
func (*SuppressRequest) ProtoMessage()               {}
func (*SuppressRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *SuppressRequest) GetSecret() string {
	if m != nil {
		return m.Secret
	}
	return ""
}

func (m *SuppressRequest) GetTimestamp() uint32 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

type Connection struct {
	DeviceName string `protobuf:"bytes,1,opt,name=DeviceName" json:"DeviceName,omitempty"`
}

func (m *Connection) Reset()                    { *m = Connection{} }
func (m *Connection) String() string            { return proto.CompactTextString(m) }
func (*Connection) ProtoMessage()               {}
func (*Connection) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *Connection) GetDeviceName() string {
	if m != nil {
		return m.DeviceName
	}
	return ""
}

type ConnectionList struct {
	Connections []*Connection `protobuf:"bytes,1,rep,name=Connections" json:"Connections,omitempty"`
}

func (m *ConnectionList) Reset()                    { *m = ConnectionList{} }
func (m *ConnectionList) String() string            { return proto.CompactTextString(m) }
func (*ConnectionList) ProtoMessage()               {}
func (*ConnectionList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *ConnectionList) GetConnections() []*Connection {
	if m != nil {
		return m.Connections
	}
	return nil
}

type WifiInfoRequest struct {
	DeviceID string   `protobuf:"bytes,1,opt,name=DeviceID" json:"DeviceID,omitempty"`
	Key      string   `protobuf:"bytes,2,opt,name=Key" json:"Key,omitempty"`
	APs      []string `protobuf:"bytes,3,rep,name=APs" json:"APs,omitempty"`
}

func (m *WifiInfoRequest) Reset()                    { *m = WifiInfoRequest{} }
func (m *WifiInfoRequest) String() string            { return proto.CompactTextString(m) }
func (*WifiInfoRequest) ProtoMessage()               {}
func (*WifiInfoRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *WifiInfoRequest) GetDeviceID() string {
	if m != nil {
		return m.DeviceID
	}
	return ""
}

func (m *WifiInfoRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *WifiInfoRequest) GetAPs() []string {
	if m != nil {
		return m.APs
	}
	return nil
}

type ButtonPressedRequest struct {
	DeviceID string `protobuf:"bytes,1,opt,name=DeviceID" json:"DeviceID,omitempty"`
	Key      string `protobuf:"bytes,2,opt,name=Key" json:"Key,omitempty"`
	ButtonID uint32 `protobuf:"varint,3,opt,name=ButtonID" json:"ButtonID,omitempty"`
}

func (m *ButtonPressedRequest) Reset()                    { *m = ButtonPressedRequest{} }
func (m *ButtonPressedRequest) String() string            { return proto.CompactTextString(m) }
func (*ButtonPressedRequest) ProtoMessage()               {}
func (*ButtonPressedRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *ButtonPressedRequest) GetDeviceID() string {
	if m != nil {
		return m.DeviceID
	}
	return ""
}

func (m *ButtonPressedRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *ButtonPressedRequest) GetButtonID() uint32 {
	if m != nil {
		return m.ButtonID
	}
	return 0
}

type Info struct {
	Text string `protobuf:"bytes,1,opt,name=Text" json:"Text,omitempty"`
}

func (m *Info) Reset()                    { *m = Info{} }
func (m *Info) String() string            { return proto.CompactTextString(m) }
func (*Info) ProtoMessage()               {}
func (*Info) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *Info) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func init() {
	proto.RegisterType((*ViaSMSRequest)(nil), "cnwnotification.ViaSMSRequest")
	proto.RegisterType((*ConfigRequest)(nil), "cnwnotification.ConfigRequest")
	proto.RegisterType((*ConfigResponse)(nil), "cnwnotification.ConfigResponse")
	proto.RegisterType((*Notification)(nil), "cnwnotification.Notification")
	proto.RegisterType((*DisplayRequest)(nil), "cnwnotification.DisplayRequest")
	proto.RegisterType((*SoundRequest)(nil), "cnwnotification.SoundRequest")
	proto.RegisterType((*SuppressRequest)(nil), "cnwnotification.SuppressRequest")
	proto.RegisterType((*Connection)(nil), "cnwnotification.Connection")
	proto.RegisterType((*ConnectionList)(nil), "cnwnotification.ConnectionList")
	proto.RegisterType((*WifiInfoRequest)(nil), "cnwnotification.WifiInfoRequest")
	proto.RegisterType((*ButtonPressedRequest)(nil), "cnwnotification.ButtonPressedRequest")
	proto.RegisterType((*Info)(nil), "cnwnotification.Info")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for CNWNotificationService service

type CNWNotificationServiceClient interface {
	GetConfig(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*ConfigResponse, error)
	Display(ctx context.Context, in *DisplayRequest, opts ...grpc.CallOption) (*common.Void, error)
	Sound(ctx context.Context, in *SoundRequest, opts ...grpc.CallOption) (*common.Void, error)
	Suppress(ctx context.Context, in *SuppressRequest, opts ...grpc.CallOption) (*common.Void, error)
	// send an SMS message through a device (if device supports it)
	ViaSMS(ctx context.Context, in *ViaSMSRequest, opts ...grpc.CallOption) (*common.Void, error)
	GetConnections(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*ConnectionList, error)
	// submit current wifi APs in range
	SendWifiInfo(ctx context.Context, in *WifiInfoRequest, opts ...grpc.CallOption) (*common.Void, error)
	GetInfo(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*Info, error)
	// notify about a change in 'info' text, so screen refreshes quicker
	NotifyInfoChange(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*common.Void, error)
	// button pressed, sent by app
	ButtonPressed(ctx context.Context, in *ButtonPressedRequest, opts ...grpc.CallOption) (*common.Void, error)
}

type cNWNotificationServiceClient struct {
	cc *grpc.ClientConn
}

func NewCNWNotificationServiceClient(cc *grpc.ClientConn) CNWNotificationServiceClient {
	return &cNWNotificationServiceClient{cc}
}

func (c *cNWNotificationServiceClient) GetConfig(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*ConfigResponse, error) {
	out := new(ConfigResponse)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/GetConfig", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cNWNotificationServiceClient) Display(ctx context.Context, in *DisplayRequest, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/Display", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cNWNotificationServiceClient) Sound(ctx context.Context, in *SoundRequest, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/Sound", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cNWNotificationServiceClient) Suppress(ctx context.Context, in *SuppressRequest, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/Suppress", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cNWNotificationServiceClient) ViaSMS(ctx context.Context, in *ViaSMSRequest, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/ViaSMS", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cNWNotificationServiceClient) GetConnections(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*ConnectionList, error) {
	out := new(ConnectionList)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/GetConnections", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cNWNotificationServiceClient) SendWifiInfo(ctx context.Context, in *WifiInfoRequest, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/SendWifiInfo", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cNWNotificationServiceClient) GetInfo(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*Info, error) {
	out := new(Info)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/GetInfo", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cNWNotificationServiceClient) NotifyInfoChange(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/NotifyInfoChange", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cNWNotificationServiceClient) ButtonPressed(ctx context.Context, in *ButtonPressedRequest, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/cnwnotification.CNWNotificationService/ButtonPressed", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for CNWNotificationService service

type CNWNotificationServiceServer interface {
	GetConfig(context.Context, *ConfigRequest) (*ConfigResponse, error)
	Display(context.Context, *DisplayRequest) (*common.Void, error)
	Sound(context.Context, *SoundRequest) (*common.Void, error)
	Suppress(context.Context, *SuppressRequest) (*common.Void, error)
	// send an SMS message through a device (if device supports it)
	ViaSMS(context.Context, *ViaSMSRequest) (*common.Void, error)
	GetConnections(context.Context, *common.Void) (*ConnectionList, error)
	// submit current wifi APs in range
	SendWifiInfo(context.Context, *WifiInfoRequest) (*common.Void, error)
	GetInfo(context.Context, *common.Void) (*Info, error)
	// notify about a change in 'info' text, so screen refreshes quicker
	NotifyInfoChange(context.Context, *common.Void) (*common.Void, error)
	// button pressed, sent by app
	ButtonPressed(context.Context, *ButtonPressedRequest) (*common.Void, error)
}

func RegisterCNWNotificationServiceServer(s *grpc.Server, srv CNWNotificationServiceServer) {
	s.RegisterService(&_CNWNotificationService_serviceDesc, srv)
}

func _CNWNotificationService_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).GetConfig(ctx, req.(*ConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CNWNotificationService_Display_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DisplayRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).Display(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/Display",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).Display(ctx, req.(*DisplayRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CNWNotificationService_Sound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SoundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).Sound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/Sound",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).Sound(ctx, req.(*SoundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CNWNotificationService_Suppress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SuppressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).Suppress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/Suppress",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).Suppress(ctx, req.(*SuppressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CNWNotificationService_ViaSMS_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ViaSMSRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).ViaSMS(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/ViaSMS",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).ViaSMS(ctx, req.(*ViaSMSRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CNWNotificationService_GetConnections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(common.Void)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).GetConnections(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/GetConnections",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).GetConnections(ctx, req.(*common.Void))
	}
	return interceptor(ctx, in, info, handler)
}

func _CNWNotificationService_SendWifiInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WifiInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).SendWifiInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/SendWifiInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).SendWifiInfo(ctx, req.(*WifiInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CNWNotificationService_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(common.Void)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/GetInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).GetInfo(ctx, req.(*common.Void))
	}
	return interceptor(ctx, in, info, handler)
}

func _CNWNotificationService_NotifyInfoChange_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(common.Void)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).NotifyInfoChange(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/NotifyInfoChange",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).NotifyInfoChange(ctx, req.(*common.Void))
	}
	return interceptor(ctx, in, info, handler)
}

func _CNWNotificationService_ButtonPressed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ButtonPressedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CNWNotificationServiceServer).ButtonPressed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cnwnotification.CNWNotificationService/ButtonPressed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CNWNotificationServiceServer).ButtonPressed(ctx, req.(*ButtonPressedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _CNWNotificationService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cnwnotification.CNWNotificationService",
	HandlerType: (*CNWNotificationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConfig",
			Handler:    _CNWNotificationService_GetConfig_Handler,
		},
		{
			MethodName: "Display",
			Handler:    _CNWNotificationService_Display_Handler,
		},
		{
			MethodName: "Sound",
			Handler:    _CNWNotificationService_Sound_Handler,
		},
		{
			MethodName: "Suppress",
			Handler:    _CNWNotificationService_Suppress_Handler,
		},
		{
			MethodName: "ViaSMS",
			Handler:    _CNWNotificationService_ViaSMS_Handler,
		},
		{
			MethodName: "GetConnections",
			Handler:    _CNWNotificationService_GetConnections_Handler,
		},
		{
			MethodName: "SendWifiInfo",
			Handler:    _CNWNotificationService_SendWifiInfo_Handler,
		},
		{
			MethodName: "GetInfo",
			Handler:    _CNWNotificationService_GetInfo_Handler,
		},
		{
			MethodName: "NotifyInfoChange",
			Handler:    _CNWNotificationService_NotifyInfoChange_Handler,
		},
		{
			MethodName: "ButtonPressed",
			Handler:    _CNWNotificationService_ButtonPressed_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "golang.conradwood.net/apis/cnwnotification/cnwnotification.proto",
}

func init() {
	proto.RegisterFile("golang.conradwood.net/apis/cnwnotification/cnwnotification.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 710 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x55, 0x6d, 0x4f, 0xdb, 0x3c,
	0x14, 0x55, 0x69, 0x29, 0xe5, 0x42, 0x00, 0x59, 0x3c, 0x28, 0xea, 0xb3, 0xb1, 0x2a, 0xda, 0x34,
	0x34, 0xa1, 0x20, 0xb1, 0x69, 0x9a, 0x26, 0xed, 0x85, 0xb6, 0x1a, 0xaa, 0x56, 0x0a, 0x24, 0xbc,
	0x48, 0xfb, 0xb4, 0x90, 0xba, 0x99, 0x45, 0x63, 0x67, 0xb1, 0x03, 0xe3, 0xeb, 0xfe, 0xf1, 0xfe,
	0xc1, 0x64, 0xc7, 0x69, 0x93, 0xb4, 0x5b, 0xa5, 0x7d, 0x4a, 0xee, 0xf1, 0xbd, 0xc7, 0xc7, 0xd7,
	0xc7, 0x36, 0x7c, 0x0c, 0xd8, 0xd8, 0xa3, 0x81, 0xed, 0x33, 0x1a, 0x7b, 0xc3, 0x7b, 0xc6, 0x86,
	0x36, 0xc5, 0xe2, 0xc0, 0x8b, 0x08, 0x3f, 0xf0, 0xe9, 0x3d, 0x65, 0x82, 0x8c, 0x88, 0xef, 0x09,
	0xc2, 0x68, 0x39, 0xb6, 0xa3, 0x98, 0x09, 0x86, 0x36, 0x4b, 0x70, 0xd3, 0xfe, 0x1b, 0x25, 0x0b,
	0x43, 0xc9, 0xa4, 0x3e, 0x29, 0x81, 0x75, 0x04, 0xc6, 0x15, 0xf1, 0xdc, 0x13, 0xd7, 0xc1, 0xdf,
	0x13, 0xcc, 0x05, 0xda, 0x81, 0xfa, 0x20, 0x09, 0x6f, 0x70, 0x6c, 0x56, 0x5a, 0x95, 0xbd, 0x55,
	0x47, 0x47, 0xc8, 0x84, 0x95, 0x13, 0xcc, 0xb9, 0x17, 0x60, 0x73, 0x49, 0x0d, 0x64, 0xa1, 0xf5,
	0x1c, 0x8c, 0x0e, 0xa3, 0x23, 0x12, 0xe4, 0x28, 0x5c, 0xec, 0xc7, 0x58, 0x64, 0x14, 0x69, 0x64,
	0x3d, 0xc0, 0x46, 0x96, 0xc8, 0x23, 0x46, 0x39, 0x46, 0x1d, 0x30, 0x06, 0x39, 0xf5, 0xdc, 0xac,
	0xb4, 0xaa, 0x7b, 0x6b, 0x87, 0x8f, 0xed, 0xf2, 0x6a, 0xf3, 0x59, 0x4e, 0xb1, 0x06, 0x6d, 0x41,
	0xf5, 0x13, 0x63, 0x5a, 0x95, 0xfc, 0x95, 0x48, 0xdb, 0x8b, 0xcd, 0x6a, 0x8a, 0xb4, 0xbd, 0xd8,
	0xba, 0x83, 0xf5, 0x7c, 0x11, 0xda, 0x86, 0xe5, 0x0b, 0x22, 0xc6, 0x58, 0x2b, 0x4c, 0x03, 0x84,
	0xa0, 0x76, 0x81, 0x7f, 0x08, 0x4d, 0xa5, 0xfe, 0x25, 0xd7, 0xa5, 0xd3, 0xcf, 0xb8, 0x2e, 0x9d,
	0xbe, 0xac, 0x75, 0x59, 0x42, 0x87, 0x66, 0x2d, 0xad, 0x55, 0x81, 0x44, 0x7b, 0xa1, 0xec, 0xce,
	0x72, 0x8a, 0xaa, 0xc0, 0xfa, 0x59, 0x81, 0x8d, 0x2e, 0xe1, 0xd1, 0xd8, 0x7b, 0x58, 0xd0, 0x1d,
	0xd9, 0x60, 0x9d, 0x99, 0x35, 0x58, 0x87, 0x68, 0x17, 0xa0, 0xed, 0xf9, 0xb7, 0x41, 0xac, 0x66,
	0x4d, 0x95, 0xe4, 0x10, 0x39, 0x2e, 0xa5, 0x76, 0xd8, 0x98, 0x25, 0xb1, 0x56, 0x95, 0x43, 0xe4,
	0xe2, 0x95, 0xc6, 0x45, 0x0a, 0xf4, 0x52, 0x97, 0xa6, 0x4b, 0xdd, 0x81, 0xfa, 0x69, 0x4c, 0x02,
	0x42, 0xf5, 0xac, 0x3a, 0x42, 0x4f, 0xc1, 0x70, 0x6f, 0x49, 0x74, 0x9e, 0x10, 0x2c, 0xbe, 0x30,
	0x8a, 0xd5, 0xa4, 0x0d, 0xa7, 0x08, 0x5a, 0xc7, 0xb0, 0xe9, 0x26, 0x51, 0x14, 0x63, 0xce, 0x17,
	0x4d, 0xfd, 0x08, 0x56, 0x2f, 0x48, 0x88, 0xb9, 0xf0, 0xc2, 0x48, 0x09, 0x30, 0x9c, 0x29, 0x60,
	0xed, 0x03, 0x74, 0x18, 0xa5, 0xd8, 0x57, 0x7b, 0xb7, 0x0b, 0xd0, 0xc5, 0x77, 0xc4, 0xc7, 0x03,
	0x2f, 0xcc, 0x36, 0x30, 0x87, 0x58, 0xa7, 0xca, 0x66, 0x3a, 0xbb, 0x4f, 0xb8, 0x40, 0xef, 0x60,
	0x6d, 0x8a, 0x64, 0x26, 0xfb, 0x7f, 0xc6, 0x64, 0xd3, 0x1c, 0x27, 0x9f, 0x6f, 0x9d, 0xc3, 0xe6,
	0x35, 0x19, 0x91, 0x1e, 0x1d, 0xb1, 0x6c, 0x1d, 0x4d, 0x68, 0xa4, 0x33, 0xf6, 0xba, 0x5a, 0xc1,
	0x24, 0x96, 0x6d, 0xfc, 0x8c, 0xb3, 0x4d, 0x94, 0xbf, 0x12, 0x39, 0x3a, 0xe3, 0x66, 0xb5, 0x55,
	0x95, 0xc8, 0xd1, 0x19, 0xb7, 0xbe, 0xc2, 0x76, 0x3b, 0x11, 0x82, 0xd1, 0x33, 0xd9, 0x1d, 0x3c,
	0xfc, 0x37, 0xde, 0x26, 0x34, 0x52, 0x96, 0x5e, 0x57, 0x6d, 0x90, 0xe1, 0x4c, 0x62, 0xab, 0x09,
	0x35, 0x29, 0x78, 0xe2, 0xe9, 0xca, 0xd4, 0xd3, 0x87, 0xbf, 0x6a, 0xb0, 0xd3, 0x19, 0x5c, 0xe7,
	0x4f, 0x84, 0x8b, 0x63, 0x39, 0x0d, 0xea, 0xc3, 0xea, 0x31, 0x16, 0xe9, 0x31, 0x45, 0xbb, 0xf3,
	0x5a, 0x34, 0x3d, 0xe8, 0xcd, 0x27, 0x7f, 0x1c, 0xd7, 0xe7, 0xfb, 0xcd, 0xc4, 0xd3, 0x68, 0x36,
	0xb7, 0x78, 0x2e, 0x9a, 0xeb, 0xb6, 0xbe, 0x98, 0xae, 0x18, 0x19, 0xa2, 0x57, 0xfa, 0x90, 0xa1,
	0xd9, 0xbb, 0x20, 0xef, 0xe5, 0x52, 0xd5, 0x5b, 0x68, 0x64, 0x8e, 0x43, 0xad, 0xd9, 0xc2, 0xa2,
	0x19, 0x4b, 0xb5, 0xaf, 0xa1, 0x9e, 0xde, 0x84, 0x73, 0x96, 0x5d, 0xb8, 0x22, 0x4b, 0x75, 0x1f,
	0x60, 0x23, 0xed, 0x58, 0xe6, 0x17, 0x54, 0x18, 0x9f, 0xdf, 0xa4, 0xbc, 0x3b, 0xdf, 0xc3, 0xba,
	0x8b, 0xe9, 0x30, 0xb3, 0xd8, 0x1c, 0xe1, 0x25, 0xf7, 0x95, 0x04, 0xd8, 0xb0, 0x72, 0x8c, 0x85,
	0x2a, 0x2d, 0xce, 0xfc, 0xdf, 0x0c, 0x91, 0x4a, 0xb2, 0x61, 0x4b, 0xed, 0xfc, 0x83, 0x8c, 0x3a,
	0xdf, 0x3c, 0x1a, 0xe0, 0x52, 0x61, 0x91, 0xbf, 0x0b, 0x46, 0xc1, 0xab, 0xe8, 0xd9, 0x0c, 0xef,
	0x3c, 0x2f, 0x17, 0x59, 0xda, 0xfb, 0xf0, 0x82, 0x62, 0x91, 0x7f, 0x97, 0xf4, 0x4b, 0x25, 0x9f,
	0xa6, 0x32, 0xe1, 0x4d, 0x5d, 0xbd, 0x4e, 0x2f, 0x7f, 0x07, 0x00, 0x00, 0xff, 0xff, 0x38, 0x7c,
	0x6e, 0x99, 0x22, 0x07, 0x00, 0x00,
}
