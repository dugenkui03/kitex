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
	"time"

	"golang.org/x/sync/singleflight"
)

var (
	// insert, not delete
	sharedTickers    sync.Map
	sharedTickersSfg singleflight.Group
)

// shared ticker
type sharedTicker struct {
	sync.Mutex
	started  bool
	interval time.Duration
	tasks    map[*Balancer]struct{}
	stopChan chan struct{}
}

func getSharedTicker(b *Balancer, refreshInterval time.Duration) *sharedTicker {
	sti, ok := sharedTickers.Load(refreshInterval)
	if ok {
		st := sti.(*sharedTicker)
		// note
		st.add(b)
		return st
	}
	v, _, _ := sharedTickersSfg.Do(refreshInterval.String(), func() (interface{}, error) {
		st := &sharedTicker{
			interval: refreshInterval,
			tasks:    map[*Balancer]struct{}{},
			stopChan: make(chan struct{}, 1),
		}
		sharedTickers.Store(refreshInterval, st)
		return st, nil
	})
	st := v.(*sharedTicker)
	// add without singleflight,
	// because we need all balancers those call this function to add themself to sharedTicker
	// note
	st.add(b)
	return st
}

func (t *sharedTicker) add(b *Balancer) {
	// note 添加任务
	t.Lock()
	defer t.Unlock()
	t.tasks[b] = struct{}{}

	// 如果ticker没有开始，则开始
	if !t.started {
		t.started = true
		go t.tick(t.interval)
	}
}

func (t *sharedTicker) delete(b *Balancer) {
	// note delete from tasks 从任务队列中删除数据
	t.Lock()
	defer t.Unlock()

	delete(t.tasks, b)
	// no tasks remaining then stop the tick
	// note 如果没有数据了，则停止 ticker
	if len(t.tasks) == 0 {
		// unblocked when multi delete call
		select {
		case t.stopChan <- struct{}{}:
			t.started = false
		default:
		}
	}
}

//tick 定时进行 负载平衡
func (t *sharedTicker) tick(interval time.Duration) {
	var wg sync.WaitGroup

	//
	// type Ticker struct {
	//	C <-chan Time // The channel on which the ticks are delivered.
	//	r runtimeTimer
	//}
	// interval 是触发任务的周期
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		// note 重要：for 循环走到这里的时候会判断两个 case 是否成立
		select {
		case <-ticker.C:
			// todo 为什么要加锁
			t.Lock()
			// map[*Balancer]struct{}， rang map的时候如果只有一个参数、则为key
			for b := range t.tasks {
				wg.Add(1)
				go func(b *Balancer) {
					defer wg.Done()
					// todo 如果 refresh() panic、会不会导致 t.Unlock() 永远不会被调用
					b.refresh()
				}(b)
			}
			wg.Wait()
			t.Unlock()
			// note 如果接收到停止的信息
		case <-t.stopChan:
			return
		}
	}
}
