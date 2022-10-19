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

// Package transport provides predefined transport protocol.
// note 返回预先定义好的传输协议
package transport

// Protocol indicates the transport protocol.
type Protocol int

// Predefined transport protocols.
// Framed is suggested.
const (
	PurePayload Protocol = 0

	TTHeader Protocol = 1 << iota // note iota 在此处是 2
	Framed                        // note 二进制协议
	HTTP
	GRPC

	TTHeaderFramed = TTHeader | Framed
)

// Unknown indicates the protocol is unknown.
const Unknown = "Unknown"

// String prints human readable information.
func (tp Protocol) String() string {
	switch tp {
	case PurePayload:
		return "PurePayload"
	case TTHeader:
		return "TTHeader"
	case Framed:
		return "Framed"
	case HTTP:
		return "HTTP"
	case TTHeaderFramed:
		return "TTHeaderFramed"
	case GRPC:
		return "GRPC"
	}
	return Unknown
}

// WithMeta reports whether the protocol is capable of carrying meta
// data besides the payload.
// note 协议是否可以携带出了元数据之外的 payload
// todo 万一是 Unknow 呢
func (tp Protocol) WithMeta() bool {
	return tp != PurePayload && tp != Framed
}
