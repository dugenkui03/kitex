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

package invoke

import (
	"errors"
	"net"
	"time"

	"github.com/cloudwego/netpoll"

	"github.com/cloudwego/kitex/pkg/remote"
	internal_netpoll "github.com/cloudwego/kitex/pkg/remote/trans/netpoll"
)

var _ Message = &message{}

// PayloadHandler is used to operate the payload.
//		note 用来操作元数据
type PayloadHandler interface {
	// note 设置请求数据、相应数据，请求数据读取缓存、相应数据读取缓存
	SetRequestBytes(buf []byte) error
	GetResponseBytes() ([]byte, error)
	GetRequestReaderByteBuffer() remote.ByteBuffer
	GetResponseWriterByteBuffer() remote.ByteBuffer
	Release() error
}

// Message is the core abstraction.
// note 核心抽象 保存了 connect信息 和 数据操作能力
type Message interface {
	// note 两个内嵌类
	net.Conn
	PayloadHandler
}

// note Message的唯一实现：本地地址、被调服务地址、请求和相应
type message struct {
	localAddr  net.Addr
	remoteAddr net.Addr
	request    remote.ByteBuffer
	response   remote.ByteBuffer
}

// NewMessage creates a new Message using the given net.addr.
func NewMessage(local, remote net.Addr) Message {
	return &message{
		localAddr:  local,
		remoteAddr: remote,
	}
}

// Read implements the Message interface.
func (f *message) Read(b []byte) (n int, err error) {
	return 0, nil
}

// Write implements the Message interface.
func (f *message) Write(b []byte) (n int, err error) {
	return 0, nil
}

// Close implements the Message interface.
func (f *message) Close() error {
	return nil
}

// LocalAddr implements the Message interface.
func (f *message) LocalAddr() net.Addr {
	return f.localAddr
}

// RemoteAddr implements the Message interface.
func (f *message) RemoteAddr() net.Addr {
	return f.remoteAddr
}

// SetDeadline implements the Message interface.
func (f *message) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline implements the Message interface.
func (f *message) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline implements the Message interface.
func (f *message) SetWriteDeadline(t time.Time) error {
	return nil
}

// SetRequestBytes implements the Message interface.
// note 设置请求数据
func (f *message) SetRequestBytes(buf []byte) error {
	f.request = remote.NewReaderBuffer(buf)
	return nil
}

// GetResponseBytes implements the Message interface.
func (f *message) GetResponseBytes() ([]byte, error) {
	if f.response == nil {
		return nil, errors.New("response not init yet")
	}
	// 返回数据、但不推进读取器
	return f.response.Peek(f.response.ReadableLen())
}

// GetRequestReaderByteBuffer implements the Message interface.
func (f *message) GetRequestReaderByteBuffer() remote.ByteBuffer {
	return f.request
}

// GetResponseWriterByteBuffer implements the Message interface.
func (f *message) GetResponseWriterByteBuffer() remote.ByteBuffer {
	if f.response == nil {
		f.response = internal_netpoll.NewReaderWriterByteBuffer(netpoll.NewLinkBuffer(0))
	}
	return f.response
}

// Release implements the Message interface.
func (f *message) Release() error {
	if f.request != nil {
		f.request.Release(nil)
	}
	if f.response != nil {
		f.response.Release(nil)
	}
	return nil
}
