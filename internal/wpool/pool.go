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

package wpool

/* Difference between wpool and gopool(github.com/bytedance/gopkg/util/gopool):
- wpool is a goroutine pool with high reuse rate. The old goroutine will block to wait for new tasks coming.
- gopool sometimes will have a very low reuse rate. The old goroutine will not block if no new task coming.
*/

import (
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

// note worker 执行的任务抽象
// Task is the function that the worker will execute.
type Task func()

// note ？工作池？绑定一些闲置的 goroutine
// Pool is a worker pool bind with some idle goroutines.
type Pool struct {
	size int32
	// note 管道，可以向里边添加元素
	tasks chan Task

	// note 工作池中可以存放的最多的闲置 worker 数量
	// maxIdle is the number of the max idle workers in the pool.
	// if maxIdle too small, the pool works like a native 'go func()'.
	maxIdle int32

	// note worker 等待任务的时间
	// maxIdleTime is the max idle time that the worker will wait for the new task.
	maxIdleTime time.Duration
}

// New creates a new worker pool.
func New(maxIdle int, maxIdleTime time.Duration) *Pool {
	return &Pool{
		tasks:       make(chan Task),
		maxIdle:     int32(maxIdle),
		maxIdleTime: maxIdleTime,
	}
}

// Size returns the number of the running workers.
func (p *Pool) Size() int32 {
	return atomic.LoadInt32(&p.size)
}

// Go creates/reuses a worker to run task.
func (p *Pool) Go(task Task) {
	select {
	case p.tasks <- task:
		// note 如果可以写入管道则直接返回
		// reuse exist worker
		return
	default:
	}

	// note 不可以写入管道则
	// create new worker
	atomic.AddInt32(&p.size, 1)
	go func() {
		// 延迟函数
		defer func() {
			if r := recover(); r != nil {
				// note 错误处理函数
				klog.Errorf("panic in wpool: error=%v: stack=%s", r, debug.Stack())
			}
			// 函数执行完后
			atomic.AddInt32(&p.size, -1)
		}()
		// note 异步执行任务
		task()
		if atomic.LoadInt32(&p.size) > p.maxIdle {
			return
		}

		// note time.NewTimer() 和 time.sleep()的区别
		// The difference is the function call time.
		//Sleep(d) will let the current goroutine enter sleeping sub-state,
		//but still stay in running state, whereas, the channel receive operation <-time.
		//After(d) will let the current goroutine enter blocking state.

		// waiting for new task
		idleTimer := time.NewTimer(p.maxIdleTime)
		for {
			select {
			case task = <-p.tasks:
				task()
				// <-idelTime.C 将当前时间阻塞timeLong：time.NewTimer(timeLong)时长
			case <-idleTimer.C:
				// worker exits
				return
			}

			// 停止阻塞行为
			if !idleTimer.Stop() {
				<-idleTimer.C
			}
			// 充值阻塞时间
			idleTimer.Reset(p.maxIdleTime)
		}
	}()
}
