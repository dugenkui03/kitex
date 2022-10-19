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

package event

import (
	"context"
	"reflect"
	"sync"

	"github.com/cloudwego/kitex/pkg/gofunc"
)

// Callback is called when the subscribed event happens.
// note 订阅事件发生时、调用此方法
//type Event struct {
//	Name   string
//	Time   time.Time
//	Detail string
//	Extra  interface{}
//}
type Callback func(*Event)

// Bus implements the observer pattern to allow event dispatching and watching.
type Bus interface {
	Watch(event string, callback Callback)
	Unwatch(event string, callback Callback)
	// note 调度
	Dispatch(event *Event)
	// note 调度并等待
	DispatchAndWait(event *Event)
}

// NewEventBus creates a Bus.
//		bus 的工厂方法
func NewEventBus() Bus {
	return &bus{}
}

// note 实现了 Bus
type bus struct {
	// note 线程安全的 Map
	callbacks sync.Map
}

// Watch subscribe to a certain event with a callback.
// note 使用回调方法订阅处理某个事件
// todo Watch 不是线程安全的
func (b *bus) Watch(event string, callback Callback) {
	var callbacks []Callback
	if actual, ok := b.callbacks.Load(event); ok {
		// note 将 actual 拼接到 callbacks、好处是不修改原来的值
		callbacks = append(callbacks, actual.([]Callback)...)
	}
	// note 添加新值
	callbacks = append(callbacks, callback)
	// note 更新map
	b.callbacks.Store(event, callbacks)
}

// Unwatch remove the given callback from the callback chain of an event.
func (b *bus) Unwatch(event string, callback Callback) {
	var filtered []Callback
	// In go, functions are not comparable, so we use reflect.ValueOf(callback).Pointer() to reflect their address for comparison.
	target := reflect.ValueOf(callback).Pointer()
	if actual, ok := b.callbacks.Load(event); ok {
		for _, h := range actual.([]Callback) {
			if reflect.ValueOf(h).Pointer() != target {
				filtered = append(filtered, h)
			}
		}
		b.callbacks.Store(event, filtered)
	}
}

// Dispatch dispatches an event by invoking each callback asynchronously.
func (b *bus) Dispatch(event *Event) {
	if actual, ok := b.callbacks.Load(event.Name); ok {
		for _, h := range actual.([]Callback) {
			f := h // assign the value to a new variable for the closure
			gofunc.GoFunc(context.Background(), func() {
				f(event)
			})
		}
	}
}

// DispatchAndWait dispatches an event by invoking callbacks concurrently and waits for them to finish.
func (b *bus) DispatchAndWait(event *Event) {
	if actual, ok := b.callbacks.Load(event.Name); ok {
		var wg sync.WaitGroup
		for i := range actual.([]Callback) {
			h := (actual.([]Callback))[i]
			wg.Add(1)
			gofunc.GoFunc(context.Background(), func() {
				h(event)
				wg.Done()
			})
		}
		wg.Wait()
	}
}
