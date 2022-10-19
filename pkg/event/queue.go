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
	"sync"
	"sync/atomic"
)

const (
	// MaxEventNum is the default size of a event queue.
	// note event 队列默认最大值
	MaxEventNum = 200
)

// Queue is a ring to collect events.
// note 收集 event 的 ring
type Queue interface {
	Push(e *Event)
	Dump() interface{}
}

// queue implements a fixed size Queue.
type queue struct {
	// note 保存队列数据
	ring []*Event
	// note 尾节点
	tail uint32
	// note ??
	tailVersion map[uint32]*uint32
	// note 读写锁
	mu sync.RWMutex
}

// NewQueue creates a queue with the given capacity.
func NewQueue(cap int) Queue {
	q := &queue{
		// note tail 初始化为0，索引对应的位置没有值
		ring:        make([]*Event, cap),
		tailVersion: make(map[uint32]*uint32, cap),
	}
	for i := 0; i <= cap; i++ {
		t := uint32(0)
		// [1...n],&0
		q.tailVersion[uint32(i)] = &t
	}
	return q
}

// Push pushes an event to the queue.
func (q *queue) Push(e *Event) {
	for {
		old := atomic.LoadUint32(&q.tail)
		new := old + 1
		// note 如果已经超过了保存队列数据的切片的len
		if new >= uint32(len(q.ring)) {
			new = 0
		}

		// <old, value +1>
		oldV := atomic.LoadUint32(q.tailVersion[old])
		newV := oldV + 1

		// note 如果一个成功、一个失败呢？是否会发生一个成功、一个失败的情况
		// 		可以使用 if boo_exp_a { time.sleep(100) if bool_exa_b { 来模拟成功一半的情况
		if atomic.CompareAndSwapUint32(&q.tail, old, new) && atomic.CompareAndSwapUint32(q.tailVersion[old], oldV, newV) {
			// note 获取写锁
			q.mu.Lock()
			// note 重要：指向尾部、但是索引对应的值为空
			q.ring[old] = e
			q.mu.Unlock()
			break
		}
	}
}

// Dump dumps the previously pushed events out in a reversed order.
// note 以相反的顺序导出之前所有的数据
// todo 也是有问题的，如果 Push 中已经更新了 tail 值、但是没有获取锁并将数据放到 ring 中，则会导致 tag 1 出执行失败，即使有数据也会导出失败
func (q *queue) Dump() interface{} {
	results := make([]*Event, 0, len(q.ring))
	q.mu.RLock()
	defer q.mu.RUnlock()
	pos := int32(q.tail)
	for i := 0; i < len(q.ring); i++ {
		pos--
		if pos < 0 {
			pos = int32(len(q.ring) - 1)
		}

		e := q.ring[pos]
		// note tag
		if e == nil {
			return results
		}

		results = append(results, e)
	}

	return results
}
