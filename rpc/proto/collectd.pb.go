// Code generated by protoc-gen-go.
// source: collectd.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	collectd.proto

It has these top-level messages:
	DispatchValuesRequest
	DispatchValuesResponse
	QueryValuesRequest
	QueryValuesResponse
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import collectd_types "collectd.org/rpc/proto/types"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

// The arguments to DispatchValues.
type DispatchValuesRequest struct {
	ValueList *collectd_types.ValueList `protobuf:"bytes,1,opt,name=value_list,json=valueList" json:"value_list,omitempty"`
}

func (m *DispatchValuesRequest) Reset()                    { *m = DispatchValuesRequest{} }
func (m *DispatchValuesRequest) String() string            { return proto1.CompactTextString(m) }
func (*DispatchValuesRequest) ProtoMessage()               {}
func (*DispatchValuesRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *DispatchValuesRequest) GetValueList() *collectd_types.ValueList {
	if m != nil {
		return m.ValueList
	}
	return nil
}

// The response from DispatchValues.
type DispatchValuesResponse struct {
}

func (m *DispatchValuesResponse) Reset()                    { *m = DispatchValuesResponse{} }
func (m *DispatchValuesResponse) String() string            { return proto1.CompactTextString(m) }
func (*DispatchValuesResponse) ProtoMessage()               {}
func (*DispatchValuesResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// The arguments to QueryValues.
type QueryValuesRequest struct {
	// Query by the fields of the identifier. Only return values matching the
	// specified shell wildcard patterns (see fnmatch(3)). Use '*' to match
	// any value.
	Identifier *collectd_types.Identifier `protobuf:"bytes,1,opt,name=identifier" json:"identifier,omitempty"`
}

func (m *QueryValuesRequest) Reset()                    { *m = QueryValuesRequest{} }
func (m *QueryValuesRequest) String() string            { return proto1.CompactTextString(m) }
func (*QueryValuesRequest) ProtoMessage()               {}
func (*QueryValuesRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *QueryValuesRequest) GetIdentifier() *collectd_types.Identifier {
	if m != nil {
		return m.Identifier
	}
	return nil
}

// The response from QueryValues.
type QueryValuesResponse struct {
	ValueList *collectd_types.ValueList `protobuf:"bytes,1,opt,name=value_list,json=valueList" json:"value_list,omitempty"`
}

func (m *QueryValuesResponse) Reset()                    { *m = QueryValuesResponse{} }
func (m *QueryValuesResponse) String() string            { return proto1.CompactTextString(m) }
func (*QueryValuesResponse) ProtoMessage()               {}
func (*QueryValuesResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *QueryValuesResponse) GetValueList() *collectd_types.ValueList {
	if m != nil {
		return m.ValueList
	}
	return nil
}

func init() {
	proto1.RegisterType((*DispatchValuesRequest)(nil), "collectd.DispatchValuesRequest")
	proto1.RegisterType((*DispatchValuesResponse)(nil), "collectd.DispatchValuesResponse")
	proto1.RegisterType((*QueryValuesRequest)(nil), "collectd.QueryValuesRequest")
	proto1.RegisterType((*QueryValuesResponse)(nil), "collectd.QueryValuesResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for Collectd service

type CollectdClient interface {
	// Query a list of values available from collectd's value cache.
	QueryValues(ctx context.Context, in *QueryValuesRequest, opts ...grpc.CallOption) (Collectd_QueryValuesClient, error)
}

type collectdClient struct {
	cc *grpc.ClientConn
}

func NewCollectdClient(cc *grpc.ClientConn) CollectdClient {
	return &collectdClient{cc}
}

func (c *collectdClient) QueryValues(ctx context.Context, in *QueryValuesRequest, opts ...grpc.CallOption) (Collectd_QueryValuesClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Collectd_serviceDesc.Streams[0], c.cc, "/collectd.Collectd/QueryValues", opts...)
	if err != nil {
		return nil, err
	}
	x := &collectdQueryValuesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Collectd_QueryValuesClient interface {
	Recv() (*QueryValuesResponse, error)
	grpc.ClientStream
}

type collectdQueryValuesClient struct {
	grpc.ClientStream
}

func (x *collectdQueryValuesClient) Recv() (*QueryValuesResponse, error) {
	m := new(QueryValuesResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Collectd service

type CollectdServer interface {
	// Query a list of values available from collectd's value cache.
	QueryValues(*QueryValuesRequest, Collectd_QueryValuesServer) error
}

func RegisterCollectdServer(s *grpc.Server, srv CollectdServer) {
	s.RegisterService(&_Collectd_serviceDesc, srv)
}

func _Collectd_QueryValues_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(QueryValuesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CollectdServer).QueryValues(m, &collectdQueryValuesServer{stream})
}

type Collectd_QueryValuesServer interface {
	Send(*QueryValuesResponse) error
	grpc.ServerStream
}

type collectdQueryValuesServer struct {
	grpc.ServerStream
}

func (x *collectdQueryValuesServer) Send(m *QueryValuesResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _Collectd_serviceDesc = grpc.ServiceDesc{
	ServiceName: "collectd.Collectd",
	HandlerType: (*CollectdServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "QueryValues",
			Handler:       _Collectd_QueryValues_Handler,
			ServerStreams: true,
		},
	},
	Metadata: fileDescriptor0,
}

// Client API for Dispatch service

type DispatchClient interface {
	// DispatchValues sends a stream of ValueLists to the server.
	DispatchValues(ctx context.Context, opts ...grpc.CallOption) (Dispatch_DispatchValuesClient, error)
}

type dispatchClient struct {
	cc *grpc.ClientConn
}

func NewDispatchClient(cc *grpc.ClientConn) DispatchClient {
	return &dispatchClient{cc}
}

func (c *dispatchClient) DispatchValues(ctx context.Context, opts ...grpc.CallOption) (Dispatch_DispatchValuesClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Dispatch_serviceDesc.Streams[0], c.cc, "/collectd.Dispatch/DispatchValues", opts...)
	if err != nil {
		return nil, err
	}
	x := &dispatchDispatchValuesClient{stream}
	return x, nil
}

type Dispatch_DispatchValuesClient interface {
	Send(*DispatchValuesRequest) error
	CloseAndRecv() (*DispatchValuesResponse, error)
	grpc.ClientStream
}

type dispatchDispatchValuesClient struct {
	grpc.ClientStream
}

func (x *dispatchDispatchValuesClient) Send(m *DispatchValuesRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *dispatchDispatchValuesClient) CloseAndRecv() (*DispatchValuesResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(DispatchValuesResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Dispatch service

type DispatchServer interface {
	// DispatchValues sends a stream of ValueLists to the server.
	DispatchValues(Dispatch_DispatchValuesServer) error
}

func RegisterDispatchServer(s *grpc.Server, srv DispatchServer) {
	s.RegisterService(&_Dispatch_serviceDesc, srv)
}

func _Dispatch_DispatchValues_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DispatchServer).DispatchValues(&dispatchDispatchValuesServer{stream})
}

type Dispatch_DispatchValuesServer interface {
	SendAndClose(*DispatchValuesResponse) error
	Recv() (*DispatchValuesRequest, error)
	grpc.ServerStream
}

type dispatchDispatchValuesServer struct {
	grpc.ServerStream
}

func (x *dispatchDispatchValuesServer) SendAndClose(m *DispatchValuesResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *dispatchDispatchValuesServer) Recv() (*DispatchValuesRequest, error) {
	m := new(DispatchValuesRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Dispatch_serviceDesc = grpc.ServiceDesc{
	ServiceName: "collectd.Dispatch",
	HandlerType: (*DispatchServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "DispatchValues",
			Handler:       _Dispatch_DispatchValues_Handler,
			ClientStreams: true,
		},
	},
	Metadata: fileDescriptor0,
}

func init() { proto1.RegisterFile("collectd.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 247 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x4b, 0xce, 0xcf, 0xc9,
	0x49, 0x4d, 0x2e, 0x49, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x80, 0xf1, 0xa5, 0xb8,
	0x4b, 0x2a, 0x0b, 0x52, 0x8b, 0x21, 0xc2, 0x4a, 0x81, 0x5c, 0xa2, 0x2e, 0x99, 0xc5, 0x05, 0x89,
	0x25, 0xc9, 0x19, 0x61, 0x89, 0x39, 0xa5, 0xa9, 0xc5, 0x41, 0xa9, 0x85, 0x40, 0xb2, 0x44, 0xc8,
	0x82, 0x8b, 0xab, 0x0c, 0x24, 0x10, 0x9f, 0x93, 0x59, 0x5c, 0x22, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1,
	0x6d, 0x24, 0xa9, 0x07, 0x37, 0x14, 0x62, 0x06, 0x58, 0x8b, 0x0f, 0x50, 0x41, 0x10, 0x67, 0x19,
	0x8c, 0xa9, 0x24, 0xc1, 0x25, 0x86, 0x6e, 0x64, 0x71, 0x41, 0x7e, 0x5e, 0x71, 0xaa, 0x52, 0x00,
	0x97, 0x50, 0x60, 0x69, 0x6a, 0x51, 0x25, 0xaa, 0x4d, 0x56, 0x5c, 0x5c, 0x99, 0x29, 0xa9, 0x79,
	0x25, 0x99, 0x69, 0x99, 0xa9, 0x45, 0x50, 0x9b, 0xa4, 0xd0, 0x6d, 0xf2, 0x84, 0xab, 0x08, 0x42,
	0x52, 0xad, 0xe4, 0xcf, 0x25, 0x8c, 0x62, 0x22, 0xc4, 0x22, 0xf2, 0x1d, 0x6f, 0x14, 0xc1, 0xc5,
	0xe1, 0x0c, 0x55, 0x26, 0xe4, 0xc3, 0xc5, 0x8d, 0x64, 0xb8, 0x90, 0x0c, 0xc2, 0x00, 0x4c, 0x5f,
	0x48, 0xc9, 0xe2, 0x90, 0x85, 0xb8, 0xc8, 0x80, 0xd1, 0x28, 0x91, 0x8b, 0x03, 0x16, 0x2c, 0x42,
	0xa1, 0x5c, 0x7c, 0xa8, 0x41, 0x24, 0x24, 0x8f, 0xd0, 0x8e, 0x35, 0x3e, 0xa4, 0x14, 0x70, 0x2b,
	0x80, 0x58, 0xa1, 0xc1, 0xe8, 0x24, 0x11, 0x25, 0x06, 0x57, 0x94, 0x5f, 0x94, 0xae, 0x5f, 0x54,
	0x90, 0xac, 0x0f, 0x8e, 0xe6, 0x24, 0x36, 0x30, 0x65, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x01,
	0xb0, 0xce, 0x45, 0x16, 0x02, 0x00, 0x00,
}