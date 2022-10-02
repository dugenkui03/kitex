/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package serviceinfo

import (
	"context"
)

// PayloadCodec alias type
type PayloadCodec int

// PayloadCodec supported by kitex.
// note 数据格式类型：thrift 或者 protobuf
const (
	Thrift PayloadCodec = iota
	Protobuf
)

// String prints human readable information.
func (p PayloadCodec) String() string {
	switch p {
	case Thrift:
		return "Thrift"
	case Protobuf:
		return "Protobuf"
	}
	panic("unknown payload type")
}

// 泛化调用的服务和方法的名称。
const (
	// GenericService name
	GenericService = "$GenericService" // private as "$"
	// GenericMethod name
	GenericMethod = "$GenericCall"
)

// note 记录 service 的元数据信息
// ServiceInfo to record meta info of service
type ServiceInfo struct {
	// deprecated, for compatibility
	PackageName string

	// The name of the service. For generic services, it is always the constant `GenericService`.
	ServiceName string

	// HandlerType is the type value of a request handler from the generated code.
	// note 来自于 生成的代码的请求处理器的类型值
	HandlerType interface{}

	// note 记录该服务拥有的信息，对于泛化服务只有一个'GenericMethod'方法
	// Methods contains the meta information of methods supported by the service.
	// For generic service, there is only one method named by the constant `GenericMethod`.
	// note 服务支持的方法名称到方法元数据
	Methods map[string]MethodInfo

	// PayloadCodec is the codec of payload.
	// 消息协议类型
	PayloadCodec PayloadCodec

	// KiteXGenVersion is the version of command line tool 'kitex'.
	// note kitex 的类型
	KiteXGenVersion string

	// Extra is for future feature info, to avoid compatibility issue
	// as otherwise we need to add a new field in the struct
	Extra map[string]interface{}

	// GenericMethod returns a MethodInfo for the given name.
	// It is used by generic calls only.
	// note 返回指定名称的方法信息，一般被泛化调用使用
	GenericMethod func(name string) MethodInfo
}

// GetPackageName returns the PackageName.
// The return value is generally taken from the Extra field of ServiceInfo with a key
// `PackageName`. For some legacy generated code, it may come from the PackageName field.
func (i *ServiceInfo) GetPackageName() (pkg string) {
	if i.PackageName != "" {
		return i.PackageName
	}
	pkg, _ = i.Extra["PackageName"].(string)
	return
}

// MethodInfo gets MethodInfo.
func (i *ServiceInfo) MethodInfo(name string) MethodInfo {
	// 泛化调用
	if i.ServiceName == GenericService {
		if i.GenericMethod != nil {
			return i.GenericMethod(name)
		}
		return i.Methods[GenericMethod]
	}
	return i.Methods[name]
}

// note 记录方法元信息：方法句柄、参数、结果和 是否oneWay
// MethodInfo to record meta info of unary method
// note 方法句柄、参数、结果是否 oneway
type MethodInfo interface {
	Handler() MethodHandler
	NewArgs() interface{}
	NewResult() interface{}
	OneWay() bool
}

// MethodHandler is corresponding to the handler wrapper func that in generated code
type MethodHandler func(ctx context.Context, handler, args, result interface{}) error

// NewMethodInfo is called in generated code to build method info
// note 使用方法句柄、参数和结果方法、isOneWay 构造 MethodInfo
func NewMethodInfo(methodHandler MethodHandler, newArgsFunc, newResultFunc func() interface{}, oneWay bool) MethodInfo {
	return methodInfo{
		handler:       methodHandler,
		newArgsFunc:   newArgsFunc,
		newResultFunc: newResultFunc,
		oneWay:        oneWay,
	}
}

type methodInfo struct {
	handler       MethodHandler
	newArgsFunc   func() interface{}
	newResultFunc func() interface{}
	oneWay        bool
}

// Handler implements the MethodInfo interface.
func (m methodInfo) Handler() MethodHandler {
	return m.handler
}

// NewArgs implements the MethodInfo interface.
func (m methodInfo) NewArgs() interface{} {
	return m.newArgsFunc()
}

// NewResult implements the MethodInfo interface.
func (m methodInfo) NewResult() interface{} {
	return m.newResultFunc()
}

// OneWay implements the MethodInfo interface.
func (m methodInfo) OneWay() bool {
	return m.oneWay
}
