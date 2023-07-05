// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: message.proto

package protocol

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// UploaderClient is the client API for Uploader service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UploaderClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
}

type uploaderClient struct {
	cc grpc.ClientConnInterface
}

func NewUploaderClient(cc grpc.ClientConnInterface) UploaderClient {
	return &uploaderClient{cc}
}

func (c *uploaderClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, "/protocol.Uploader/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UploaderServer is the server API for Uploader service.
// All implementations must embed UnimplementedUploaderServer
// for forward compatibility
type UploaderServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	mustEmbedUnimplementedUploaderServer()
}

// UnimplementedUploaderServer must be embedded to have forward compatible implementations.
type UnimplementedUploaderServer struct {
}

func (UnimplementedUploaderServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedUploaderServer) mustEmbedUnimplementedUploaderServer() {}

// UnsafeUploaderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UploaderServer will
// result in compilation errors.
type UnsafeUploaderServer interface {
	mustEmbedUnimplementedUploaderServer()
}

func RegisterUploaderServer(s grpc.ServiceRegistrar, srv UploaderServer) {
	s.RegisterService(&Uploader_ServiceDesc, srv)
}

func _Uploader_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UploaderServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.Uploader/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UploaderServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Uploader_ServiceDesc is the grpc.ServiceDesc for Uploader service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Uploader_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protocol.Uploader",
	HandlerType: (*UploaderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _Uploader_Register_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "message.proto",
}

// StorageClient is the client API for Storage service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StorageClient interface {
	CheckReadiness(ctx context.Context, in *CheckReadinessRequest, opts ...grpc.CallOption) (*CheckReadinessResponse, error)
	UploadFile(ctx context.Context, opts ...grpc.CallOption) (Storage_UploadFileClient, error)
}

type storageClient struct {
	cc grpc.ClientConnInterface
}

func NewStorageClient(cc grpc.ClientConnInterface) StorageClient {
	return &storageClient{cc}
}

func (c *storageClient) CheckReadiness(ctx context.Context, in *CheckReadinessRequest, opts ...grpc.CallOption) (*CheckReadinessResponse, error) {
	out := new(CheckReadinessResponse)
	err := c.cc.Invoke(ctx, "/protocol.Storage/CheckReadiness", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) UploadFile(ctx context.Context, opts ...grpc.CallOption) (Storage_UploadFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &Storage_ServiceDesc.Streams[0], "/protocol.Storage/UploadFile", opts...)
	if err != nil {
		return nil, err
	}
	x := &storageUploadFileClient{stream}
	return x, nil
}

type Storage_UploadFileClient interface {
	Send(*UploadFileRequest) error
	CloseAndRecv() (*UploadFileResponse, error)
	grpc.ClientStream
}

type storageUploadFileClient struct {
	grpc.ClientStream
}

func (x *storageUploadFileClient) Send(m *UploadFileRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *storageUploadFileClient) CloseAndRecv() (*UploadFileResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(UploadFileResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// StorageServer is the server API for Storage service.
// All implementations must embed UnimplementedStorageServer
// for forward compatibility
type StorageServer interface {
	CheckReadiness(context.Context, *CheckReadinessRequest) (*CheckReadinessResponse, error)
	UploadFile(Storage_UploadFileServer) error
	mustEmbedUnimplementedStorageServer()
}

// UnimplementedStorageServer must be embedded to have forward compatible implementations.
type UnimplementedStorageServer struct {
}

func (UnimplementedStorageServer) CheckReadiness(context.Context, *CheckReadinessRequest) (*CheckReadinessResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckReadiness not implemented")
}
func (UnimplementedStorageServer) UploadFile(Storage_UploadFileServer) error {
	return status.Errorf(codes.Unimplemented, "method UploadFile not implemented")
}
func (UnimplementedStorageServer) mustEmbedUnimplementedStorageServer() {}

// UnsafeStorageServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StorageServer will
// result in compilation errors.
type UnsafeStorageServer interface {
	mustEmbedUnimplementedStorageServer()
}

func RegisterStorageServer(s grpc.ServiceRegistrar, srv StorageServer) {
	s.RegisterService(&Storage_ServiceDesc, srv)
}

func _Storage_CheckReadiness_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckReadinessRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).CheckReadiness(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.Storage/CheckReadiness",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).CheckReadiness(ctx, req.(*CheckReadinessRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_UploadFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(StorageServer).UploadFile(&storageUploadFileServer{stream})
}

type Storage_UploadFileServer interface {
	SendAndClose(*UploadFileResponse) error
	Recv() (*UploadFileRequest, error)
	grpc.ServerStream
}

type storageUploadFileServer struct {
	grpc.ServerStream
}

func (x *storageUploadFileServer) SendAndClose(m *UploadFileResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *storageUploadFileServer) Recv() (*UploadFileRequest, error) {
	m := new(UploadFileRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Storage_ServiceDesc is the grpc.ServiceDesc for Storage service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Storage_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protocol.Storage",
	HandlerType: (*StorageServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckReadiness",
			Handler:    _Storage_CheckReadiness_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "UploadFile",
			Handler:       _Storage_UploadFile_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "message.proto",
}
