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

package rpcinfo

import (
	"context"
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/cloudwego/kitex/transport"
)

// note 包含 c/s 某端的信息
// EndpointInfo contains info for endpoint.
type EndpointInfo interface {
	ServiceName() string
	Method() string
	Address() net.Addr
	Tag(key string) (value string, exist bool)
	DefaultTag(key, def string) string
}

// RPCStats is used to collect statistics about the RPC.
type RPCStats interface {
	Record(ctx context.Context, event stats.Event, status stats.Status, info string)
	SendSize() uint64
	RecvSize() uint64
	Error() error
	Panicked() (bool, interface{})
	GetEvent(event stats.Event) Event
	Level() stats.Level
}

// Event is the abstraction of an event happened at a specific time.
type Event interface {
	Event() stats.Event
	Status() stats.Status
	Info() string
	Time() time.Time
	IsNil() bool
}

// Timeouts contains settings of timeouts.
// note rpc超时、连接超时、读写超时
type Timeouts interface {
	RPCTimeout() time.Duration
	ConnectTimeout() time.Duration
	ReadWriteTimeout() time.Duration
}

// TimeoutProvider provides timeout settings.
type TimeoutProvider interface {
	Timeouts(ri RPCInfo) Timeouts
}

// RPCConfig contains configuration for RPC.
// note 包含 RPC 的配置信息
type RPCConfig interface {
	//	RPCTimeout() time.Duration
	//	ConnectTimeout() time.Duration
	//	ReadWriteTimeout() time.Duration
	Timeouts

	IOBufferSize() int
	TransportProtocol() transport.Protocol
	InteractionMode() InteractionMode
}

// Invocation contains specific information about the call.
type Invocation interface {
	PackageName() string
	ServiceName() string
	MethodName() string
	SeqID() int32
}

// RPCInfo is the core abstraction of information about an RPC in Kitex.
// note kitex 中 RPC 的核心抽象
type RPCInfo interface {
	// 谁发的信息
	From() EndpointInfo
	// 谁是接受者
	To() EndpointInfo
	// 调用信息：
	Invocation() Invocation
	Config() RPCConfig
	Stats() RPCStats
}
