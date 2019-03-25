// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stellarproject/orbit/api/v1/dhcp.proto

/*
	Package v1 is a generated protocol buffer package.

	It is generated from these files:
		github.com/stellarproject/orbit/api/v1/dhcp.proto
		github.com/stellarproject/orbit/api/v1/orbit.proto

	It has these top-level messages:
		DHCPAddRequest
		DHCPAddResponse
		DHCPDeleteRequest
		CNIIP
		CNIRoute
		CNIIPNet
		CreateRequest
		DeleteRequest
		GetRequest
		GetResponse
		KillRequest
		ListRequest
		ListResponse
		ContainerInfo
		Snapshot
		RollbackRequest
		RollbackResponse
		StartRequest
		StopRequest
		UpdateRequest
		UpdateResponse
		PushRequest
		CheckpointRequest
		CheckpointResponse
		RestoreRequest
		RestoreResponse
		MigrateRequest
		MigrateResponse
		HostNetwork
		CNIIPAM
		CNINetwork
		Security
		Container
		Volume
		Config
		ServiceConfig
		Service
		HealthCheck
		GPUs
		Resources
		Mount
		Process
		User
*/
package v1

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// skipping weak import gogoproto "github.com/gogo/protobuf/gogoproto"
import google_protobuf1 "github.com/gogo/protobuf/types"

import context "golang.org/x/net/context"
import grpc "google.golang.org/grpc"

import strings "strings"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type DHCPAddRequest struct {
	ID    string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Iface string `protobuf:"bytes,2,opt,name=iface,proto3" json:"iface,omitempty"`
	Netns string `protobuf:"bytes,3,opt,name=netns,proto3" json:"netns,omitempty"`
	Name  string `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *DHCPAddRequest) Reset()                    { *m = DHCPAddRequest{} }
func (*DHCPAddRequest) ProtoMessage()               {}
func (*DHCPAddRequest) Descriptor() ([]byte, []int) { return fileDescriptorDhcp, []int{0} }

type DHCPAddResponse struct {
	IPs    []*CNIIP    `protobuf:"bytes,1,rep,name=ips" json:"ips,omitempty"`
	Routes []*CNIRoute `protobuf:"bytes,2,rep,name=routes" json:"routes,omitempty"`
}

func (m *DHCPAddResponse) Reset()                    { *m = DHCPAddResponse{} }
func (*DHCPAddResponse) ProtoMessage()               {}
func (*DHCPAddResponse) Descriptor() ([]byte, []int) { return fileDescriptorDhcp, []int{1} }

type DHCPDeleteRequest struct {
	ID    string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Iface string `protobuf:"bytes,3,opt,name=iface,proto3" json:"iface,omitempty"`
	Netns string `protobuf:"bytes,4,opt,name=netns,proto3" json:"netns,omitempty"`
}

func (m *DHCPDeleteRequest) Reset()                    { *m = DHCPDeleteRequest{} }
func (*DHCPDeleteRequest) ProtoMessage()               {}
func (*DHCPDeleteRequest) Descriptor() ([]byte, []int) { return fileDescriptorDhcp, []int{2} }

type CNIIP struct {
	Version string    `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	Address *CNIIPNet `protobuf:"bytes,2,opt,name=address" json:"address,omitempty"`
	Gateway []byte    `protobuf:"bytes,3,opt,name=gateway,proto3" json:"gateway,omitempty"`
}

func (m *CNIIP) Reset()                    { *m = CNIIP{} }
func (*CNIIP) ProtoMessage()               {}
func (*CNIIP) Descriptor() ([]byte, []int) { return fileDescriptorDhcp, []int{3} }

type CNIRoute struct {
	Dst *CNIIPNet `protobuf:"bytes,1,opt,name=dst" json:"dst,omitempty"`
	Gw  []byte    `protobuf:"bytes,2,opt,name=gw,proto3" json:"gw,omitempty"`
}

func (m *CNIRoute) Reset()                    { *m = CNIRoute{} }
func (*CNIRoute) ProtoMessage()               {}
func (*CNIRoute) Descriptor() ([]byte, []int) { return fileDescriptorDhcp, []int{4} }

type CNIIPNet struct {
	IP   []byte `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Mask []byte `protobuf:"bytes,2,opt,name=mask,proto3" json:"mask,omitempty"`
}

func (m *CNIIPNet) Reset()                    { *m = CNIIPNet{} }
func (*CNIIPNet) ProtoMessage()               {}
func (*CNIIPNet) Descriptor() ([]byte, []int) { return fileDescriptorDhcp, []int{5} }

func init() {
	proto.RegisterType((*DHCPAddRequest)(nil), "io.orbit.v1.DHCPAddRequest")
	proto.RegisterType((*DHCPAddResponse)(nil), "io.orbit.v1.DHCPAddResponse")
	proto.RegisterType((*DHCPDeleteRequest)(nil), "io.orbit.v1.DHCPDeleteRequest")
	proto.RegisterType((*CNIIP)(nil), "io.orbit.v1.CNIIP")
	proto.RegisterType((*CNIRoute)(nil), "io.orbit.v1.CNIRoute")
	proto.RegisterType((*CNIIPNet)(nil), "io.orbit.v1.CNIIPNet")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for DHCP service

type DHCPClient interface {
	DHCPAdd(ctx context.Context, in *DHCPAddRequest, opts ...grpc.CallOption) (*DHCPAddResponse, error)
	DHCPDelete(ctx context.Context, in *DHCPDeleteRequest, opts ...grpc.CallOption) (*google_protobuf1.Empty, error)
}

type dHCPClient struct {
	cc *grpc.ClientConn
}

func NewDHCPClient(cc *grpc.ClientConn) DHCPClient {
	return &dHCPClient{cc}
}

func (c *dHCPClient) DHCPAdd(ctx context.Context, in *DHCPAddRequest, opts ...grpc.CallOption) (*DHCPAddResponse, error) {
	out := new(DHCPAddResponse)
	err := grpc.Invoke(ctx, "/io.orbit.v1.DHCP/DHCPAdd", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dHCPClient) DHCPDelete(ctx context.Context, in *DHCPDeleteRequest, opts ...grpc.CallOption) (*google_protobuf1.Empty, error) {
	out := new(google_protobuf1.Empty)
	err := grpc.Invoke(ctx, "/io.orbit.v1.DHCP/DHCPDelete", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for DHCP service

type DHCPServer interface {
	DHCPAdd(context.Context, *DHCPAddRequest) (*DHCPAddResponse, error)
	DHCPDelete(context.Context, *DHCPDeleteRequest) (*google_protobuf1.Empty, error)
}

func RegisterDHCPServer(s *grpc.Server, srv DHCPServer) {
	s.RegisterService(&_DHCP_serviceDesc, srv)
}

func _DHCP_DHCPAdd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DHCPAddRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DHCPServer).DHCPAdd(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.orbit.v1.DHCP/DHCPAdd",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DHCPServer).DHCPAdd(ctx, req.(*DHCPAddRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DHCP_DHCPDelete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DHCPDeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DHCPServer).DHCPDelete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.orbit.v1.DHCP/DHCPDelete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DHCPServer).DHCPDelete(ctx, req.(*DHCPDeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _DHCP_serviceDesc = grpc.ServiceDesc{
	ServiceName: "io.orbit.v1.DHCP",
	HandlerType: (*DHCPServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DHCPAdd",
			Handler:    _DHCP_DHCPAdd_Handler,
		},
		{
			MethodName: "DHCPDelete",
			Handler:    _DHCP_DHCPDelete_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/stellarproject/orbit/api/v1/dhcp.proto",
}

func (m *DHCPAddRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DHCPAddRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.ID) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.ID)))
		i += copy(dAtA[i:], m.ID)
	}
	if len(m.Iface) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Iface)))
		i += copy(dAtA[i:], m.Iface)
	}
	if len(m.Netns) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Netns)))
		i += copy(dAtA[i:], m.Netns)
	}
	if len(m.Name) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	return i, nil
}

func (m *DHCPAddResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DHCPAddResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.IPs) > 0 {
		for _, msg := range m.IPs {
			dAtA[i] = 0xa
			i++
			i = encodeVarintDhcp(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Routes) > 0 {
		for _, msg := range m.Routes {
			dAtA[i] = 0x12
			i++
			i = encodeVarintDhcp(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *DHCPDeleteRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DHCPDeleteRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.ID) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.ID)))
		i += copy(dAtA[i:], m.ID)
	}
	if len(m.Name) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.Iface) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Iface)))
		i += copy(dAtA[i:], m.Iface)
	}
	if len(m.Netns) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Netns)))
		i += copy(dAtA[i:], m.Netns)
	}
	return i, nil
}

func (m *CNIIP) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CNIIP) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Version) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Version)))
		i += copy(dAtA[i:], m.Version)
	}
	if m.Address != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(m.Address.Size()))
		n1, err := m.Address.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if len(m.Gateway) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Gateway)))
		i += copy(dAtA[i:], m.Gateway)
	}
	return i, nil
}

func (m *CNIRoute) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CNIRoute) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Dst != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(m.Dst.Size()))
		n2, err := m.Dst.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	if len(m.Gw) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Gw)))
		i += copy(dAtA[i:], m.Gw)
	}
	return i, nil
}

func (m *CNIIPNet) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CNIIPNet) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.IP) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.IP)))
		i += copy(dAtA[i:], m.IP)
	}
	if len(m.Mask) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintDhcp(dAtA, i, uint64(len(m.Mask)))
		i += copy(dAtA[i:], m.Mask)
	}
	return i, nil
}

func encodeVarintDhcp(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *DHCPAddRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	l = len(m.Iface)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	l = len(m.Netns)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	return n
}

func (m *DHCPAddResponse) Size() (n int) {
	var l int
	_ = l
	if len(m.IPs) > 0 {
		for _, e := range m.IPs {
			l = e.Size()
			n += 1 + l + sovDhcp(uint64(l))
		}
	}
	if len(m.Routes) > 0 {
		for _, e := range m.Routes {
			l = e.Size()
			n += 1 + l + sovDhcp(uint64(l))
		}
	}
	return n
}

func (m *DHCPDeleteRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	l = len(m.Iface)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	l = len(m.Netns)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	return n
}

func (m *CNIIP) Size() (n int) {
	var l int
	_ = l
	l = len(m.Version)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	if m.Address != nil {
		l = m.Address.Size()
		n += 1 + l + sovDhcp(uint64(l))
	}
	l = len(m.Gateway)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	return n
}

func (m *CNIRoute) Size() (n int) {
	var l int
	_ = l
	if m.Dst != nil {
		l = m.Dst.Size()
		n += 1 + l + sovDhcp(uint64(l))
	}
	l = len(m.Gw)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	return n
}

func (m *CNIIPNet) Size() (n int) {
	var l int
	_ = l
	l = len(m.IP)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	l = len(m.Mask)
	if l > 0 {
		n += 1 + l + sovDhcp(uint64(l))
	}
	return n
}

func sovDhcp(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozDhcp(x uint64) (n int) {
	return sovDhcp(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *DHCPAddRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&DHCPAddRequest{`,
		`ID:` + fmt.Sprintf("%v", this.ID) + `,`,
		`Iface:` + fmt.Sprintf("%v", this.Iface) + `,`,
		`Netns:` + fmt.Sprintf("%v", this.Netns) + `,`,
		`Name:` + fmt.Sprintf("%v", this.Name) + `,`,
		`}`,
	}, "")
	return s
}
func (this *DHCPAddResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&DHCPAddResponse{`,
		`IPs:` + strings.Replace(fmt.Sprintf("%v", this.IPs), "CNIIP", "CNIIP", 1) + `,`,
		`Routes:` + strings.Replace(fmt.Sprintf("%v", this.Routes), "CNIRoute", "CNIRoute", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *DHCPDeleteRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&DHCPDeleteRequest{`,
		`ID:` + fmt.Sprintf("%v", this.ID) + `,`,
		`Name:` + fmt.Sprintf("%v", this.Name) + `,`,
		`Iface:` + fmt.Sprintf("%v", this.Iface) + `,`,
		`Netns:` + fmt.Sprintf("%v", this.Netns) + `,`,
		`}`,
	}, "")
	return s
}
func (this *CNIIP) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&CNIIP{`,
		`Version:` + fmt.Sprintf("%v", this.Version) + `,`,
		`Address:` + strings.Replace(fmt.Sprintf("%v", this.Address), "CNIIPNet", "CNIIPNet", 1) + `,`,
		`Gateway:` + fmt.Sprintf("%v", this.Gateway) + `,`,
		`}`,
	}, "")
	return s
}
func (this *CNIRoute) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&CNIRoute{`,
		`Dst:` + strings.Replace(fmt.Sprintf("%v", this.Dst), "CNIIPNet", "CNIIPNet", 1) + `,`,
		`Gw:` + fmt.Sprintf("%v", this.Gw) + `,`,
		`}`,
	}, "")
	return s
}
func (this *CNIIPNet) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&CNIIPNet{`,
		`IP:` + fmt.Sprintf("%v", this.IP) + `,`,
		`Mask:` + fmt.Sprintf("%v", this.Mask) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringDhcp(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *DHCPAddRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDhcp
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DHCPAddRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DHCPAddRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Iface", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Iface = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Netns", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Netns = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDhcp(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDhcp
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *DHCPAddResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDhcp
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DHCPAddResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DHCPAddResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IPs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IPs = append(m.IPs, &CNIIP{})
			if err := m.IPs[len(m.IPs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Routes", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Routes = append(m.Routes, &CNIRoute{})
			if err := m.Routes[len(m.Routes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDhcp(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDhcp
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *DHCPDeleteRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDhcp
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DHCPDeleteRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DHCPDeleteRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Iface", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Iface = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Netns", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Netns = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDhcp(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDhcp
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *CNIIP) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDhcp
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: CNIIP: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CNIIP: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Version = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Address == nil {
				m.Address = &CNIIPNet{}
			}
			if err := m.Address.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Gateway", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Gateway = append(m.Gateway[:0], dAtA[iNdEx:postIndex]...)
			if m.Gateway == nil {
				m.Gateway = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDhcp(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDhcp
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *CNIRoute) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDhcp
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: CNIRoute: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CNIRoute: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Dst", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Dst == nil {
				m.Dst = &CNIIPNet{}
			}
			if err := m.Dst.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Gw", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Gw = append(m.Gw[:0], dAtA[iNdEx:postIndex]...)
			if m.Gw == nil {
				m.Gw = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDhcp(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDhcp
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *CNIIPNet) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDhcp
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: CNIIPNet: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CNIIPNet: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IP", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IP = append(m.IP[:0], dAtA[iNdEx:postIndex]...)
			if m.IP == nil {
				m.IP = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Mask", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthDhcp
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Mask = append(m.Mask[:0], dAtA[iNdEx:postIndex]...)
			if m.Mask == nil {
				m.Mask = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDhcp(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDhcp
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipDhcp(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowDhcp
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowDhcp
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthDhcp
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowDhcp
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipDhcp(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthDhcp = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowDhcp   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stellarproject/orbit/api/v1/dhcp.proto", fileDescriptorDhcp)
}

var fileDescriptorDhcp = []byte{
	// 491 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x8d, 0xed, 0x34, 0x81, 0x69, 0x54, 0xc4, 0xaa, 0x54, 0x56, 0x8a, 0xdc, 0xca, 0x17, 0xca,
	0xa1, 0xb6, 0x52, 0x24, 0x2e, 0x9c, 0x48, 0x8c, 0x84, 0x2f, 0x95, 0xb5, 0x27, 0xc4, 0xcd, 0x89,
	0xb7, 0xce, 0x52, 0xc7, 0x6b, 0xbc, 0x9b, 0x44, 0xbd, 0xf1, 0x0d, 0x7c, 0x55, 0x8f, 0x1c, 0x39,
	0x55, 0xd4, 0x5f, 0x82, 0x76, 0x6c, 0x93, 0x06, 0x82, 0xe0, 0xb6, 0x33, 0xf3, 0x66, 0xde, 0xbc,
	0x99, 0x59, 0x18, 0xa5, 0x5c, 0xcd, 0x97, 0x53, 0x6f, 0x26, 0x16, 0xbe, 0x54, 0x2c, 0xcb, 0xe2,
	0xb2, 0x28, 0xc5, 0x27, 0x36, 0x53, 0xbe, 0x28, 0xa7, 0x5c, 0xf9, 0x71, 0xc1, 0xfd, 0xd5, 0xc8,
	0x4f, 0xe6, 0xb3, 0xc2, 0x2b, 0x4a, 0xa1, 0x04, 0xd9, 0xe7, 0xc2, 0xc3, 0x98, 0xb7, 0x1a, 0x0d,
	0x0f, 0x53, 0x91, 0x0a, 0xf4, 0xfb, 0xfa, 0x55, 0x43, 0x86, 0xc7, 0xa9, 0x10, 0x69, 0xc6, 0x7c,
	0xb4, 0xa6, 0xcb, 0x2b, 0x9f, 0x2d, 0x0a, 0x75, 0x53, 0x07, 0xdd, 0x39, 0x1c, 0x04, 0xef, 0x27,
	0xd1, 0xdb, 0x24, 0xa1, 0xec, 0xf3, 0x92, 0x49, 0x45, 0x8e, 0xc0, 0xe4, 0x89, 0x6d, 0x9c, 0x1a,
	0x67, 0x8f, 0xc7, 0xbd, 0xea, 0xee, 0xc4, 0x0c, 0x03, 0x6a, 0xf2, 0x84, 0x1c, 0xc2, 0x1e, 0xbf,
	0x8a, 0x67, 0xcc, 0x36, 0x75, 0x88, 0xd6, 0x86, 0xf6, 0xe6, 0x4c, 0xe5, 0xd2, 0xb6, 0x6a, 0x2f,
	0x1a, 0x84, 0x40, 0x37, 0x8f, 0x17, 0xcc, 0xee, 0xa2, 0x13, 0xdf, 0xae, 0x80, 0x27, 0xbf, 0x98,
	0x64, 0x21, 0x72, 0xc9, 0xc8, 0x39, 0x58, 0xbc, 0x90, 0xb6, 0x71, 0x6a, 0x9d, 0xed, 0x5f, 0x10,
	0xef, 0x81, 0x14, 0x6f, 0x72, 0x19, 0x86, 0xd1, 0xb8, 0x5f, 0xdd, 0x9d, 0x58, 0x61, 0x24, 0xa9,
	0xc6, 0x91, 0x73, 0xe8, 0x95, 0x62, 0xa9, 0x98, 0xb4, 0x4d, 0xcc, 0x78, 0xf6, 0x7b, 0x06, 0xd5,
	0x51, 0xda, 0x80, 0xdc, 0x6b, 0x78, 0xaa, 0x09, 0x03, 0x96, 0x31, 0xc5, 0xfe, 0xa5, 0xae, 0xed,
	0xd8, 0xdc, 0x74, 0xbc, 0x51, 0x6c, 0xed, 0x54, 0xdc, 0x7d, 0xa0, 0xd8, 0xcd, 0x60, 0x0f, 0x5b,
	0x26, 0x36, 0xf4, 0x57, 0xac, 0x94, 0x5c, 0xe4, 0x35, 0x0b, 0x6d, 0x4d, 0xe2, 0x43, 0x3f, 0x4e,
	0x92, 0x92, 0x49, 0x89, 0x2c, 0x3b, 0xfa, 0x0f, 0xa3, 0x4b, 0xa6, 0x68, 0x8b, 0xd2, 0xa5, 0xd2,
	0x58, 0xb1, 0x75, 0x7c, 0x83, 0x1d, 0x0c, 0x68, 0x6b, 0xba, 0x13, 0x78, 0xd4, 0xca, 0x25, 0x2f,
	0xc0, 0x4a, 0xa4, 0x42, 0xb2, 0xbf, 0x96, 0xd4, 0x08, 0x72, 0x00, 0x66, 0xba, 0x46, 0xea, 0x01,
	0x35, 0xd3, 0xb5, 0xfb, 0x1a, 0x8b, 0x20, 0x00, 0xc7, 0x52, 0x60, 0x8d, 0x41, 0x33, 0x96, 0x88,
	0x9a, 0xbc, 0xd0, 0x63, 0x59, 0xc4, 0xf2, 0xba, 0xc9, 0xc2, 0xf7, 0xc5, 0x57, 0x03, 0xba, 0x7a,
	0xb0, 0x24, 0x80, 0x7e, 0xb3, 0x51, 0x72, 0xbc, 0xc5, 0xbb, 0x7d, 0x51, 0xc3, 0xe7, 0xbb, 0x83,
	0xcd, 0x11, 0x04, 0x00, 0x9b, 0x35, 0x11, 0xe7, 0x0f, 0xec, 0xd6, 0xfe, 0x86, 0x47, 0x5e, 0x7d,
	0xcd, 0x5e, 0x7b, 0xcd, 0xde, 0x3b, 0x7d, 0xcd, 0xe3, 0xc9, 0xed, 0xbd, 0xd3, 0xf9, 0x7e, 0xef,
	0x74, 0xbe, 0x54, 0x8e, 0x71, 0x5b, 0x39, 0xc6, 0xb7, 0xca, 0x31, 0x7e, 0x54, 0x8e, 0xf1, 0xf1,
	0xe5, 0xff, 0x7d, 0xaa, 0x37, 0xab, 0xd1, 0x87, 0xce, 0xb4, 0x87, 0x65, 0x5f, 0xfd, 0x0c, 0x00,
	0x00, 0xff, 0xff, 0x5e, 0xec, 0x68, 0x29, 0x8a, 0x03, 0x00, 0x00,
}
