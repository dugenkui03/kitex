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

package lbcache

import (
	"sync"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/loadbalance"
)

// intercept the rebalancer and call hooks
// wrap the loadbalance.Rebalancer and execute registered hooks
// note loadbalance.Rebalancer 的切面方法
type hookableRebalancer struct {
	inner          loadbalance.Rebalancer
	rebalanceL     sync.Mutex
	rebalanceIndex int
	rebalanceHooks map[int]func(*discovery.Change)
	deleteL        sync.Mutex
	deleteIndex    int
	deleteHooks    map[int]func(*discovery.Change)
}

var (
	_ loadbalance.Rebalancer = (*hookableRebalancer)(nil)
	_ Hookable               = (*hookableRebalancer)(nil)
)

// 默认的 平衡器钩子对象
func newHookRebalancer(inner loadbalance.Rebalancer) *hookableRebalancer {
	return &hookableRebalancer{
		inner:          inner,
		rebalanceHooks: map[int]func(*discovery.Change){},
		deleteHooks:    map[int]func(*discovery.Change){},
	}
}

func (b *hookableRebalancer) Rebalance(ch discovery.Change) {
	b.rebalanceL.Lock()
	for _, h := range b.rebalanceHooks {
		h(&ch)
	}
	b.rebalanceL.Unlock()

	b.inner.Rebalance(ch)
}

func (b *hookableRebalancer) Delete(ch discovery.Change) {
	b.deleteL.Lock()
	for _, h := range b.deleteHooks {
		h(&ch)
	}
	b.deleteL.Unlock()

	b.inner.Delete(ch)
}

func (b *hookableRebalancer) RegisterRebalanceHook(f func(ch *discovery.Change)) int {
	b.rebalanceL.Lock()
	defer b.rebalanceL.Unlock()

	index := b.rebalanceIndex
	b.rebalanceIndex++
	b.rebalanceHooks[index] = f
	return index
}

func (b *hookableRebalancer) DeregisterRebalanceHook(index int) {
	b.rebalanceL.Lock()
	defer b.rebalanceL.Unlock()

	delete(b.rebalanceHooks, index)
}

func (b *hookableRebalancer) RegisterDeleteHook(f func(ch *discovery.Change)) int {
	b.deleteL.Lock()
	defer b.deleteL.Unlock()

	index := b.deleteIndex
	b.deleteIndex++
	b.deleteHooks[index] = f
	return index
}

func (b *hookableRebalancer) DeregisterDeleteHook(index int) {
	b.deleteL.Lock()
	defer b.deleteL.Unlock()

	delete(b.deleteHooks, index)
}

// 啥也不干
type noopHookRebalancer struct{}

// note Hookable 是接口、 noopHookRebalancer 是结构类型。_ 可以在编译的时候校验 noopHookRebalancer 实现了 Hookable 接口
var _ Hookable = (*noopHookRebalancer)(nil)

func (noopHookRebalancer) RegisterRebalanceHook(func(ch *discovery.Change)) int { return 0 }
func (noopHookRebalancer) RegisterDeleteHook(func(ch *discovery.Change)) int    { return 0 }
func (noopHookRebalancer) DeregisterRebalanceHook(index int)                    {}
func (noopHookRebalancer) DeregisterDeleteHook(index int)                       {}
