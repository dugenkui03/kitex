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
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	mocksloadbalance "github.com/cloudwego/kitex/internal/mocks/loadbalance"
	"github.com/cloudwego/kitex/internal/test"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

var defaultOptions = Options{
	RefreshInterval: defaultRefreshInterval,
	ExpireInterval:  defaultExpireInterval,
}

func TestBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 获取目标服务的实例
	ins := discovery.NewInstance("tcp", "127.0.0.1:8888", 10, nil)

	var _ discovery.Resolver = (*discovery.SynthesizedResolver)(nil)
	resolver := &discovery.SynthesizedResolver{

		// func(ctx context.Context, key string) (Result, error)
		// note 根据对目标端点的描述、获取目标端点的实例列表
		ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
			return discovery.Result{Cacheable: true, CacheKey: key, Instances: []discovery.Instance{ins}}, nil
		},

		// note 返回 target端点信息的描述、该描述应该适合作为缓存的key
		TargetFunc: func(ctx context.Context, target rpcinfo.EndpointInfo) string {
			return "mockRoute"
		},

		// func() string
		NameFunc: func() string {
			// note 表示返回当前 test case 的名称，即 TestBuilder
			return t.Name()
		},
	}

	// lb 类型：MockLoadbalancer
	lb := mocksloadbalance.NewMockLoadbalancer(ctrl)

	lb.EXPECT().
		// GetPicker  indicates an expected call of GetPicker
		// returns a matcher that always matches.
		GetPicker(gomock.Any()).
		DoAndReturn(
			func(res discovery.Result) loadbalance.Picker {
				test.Assert(t, res.Cacheable)
				test.Assert(t, res.CacheKey == t.Name()+":mockRoute", res.CacheKey)
				test.Assert(t, len(res.Instances) == 1)
				test.Assert(t, res.Instances[0].Address().String() == "127.0.0.1:8888")
				picker := mocksloadbalance.NewMockPicker(ctrl)
				return picker
			}).AnyTimes()

	// returns an object that allows the caller to indicate expected use
	lb.EXPECT().
		// indicates an expected call of Name.
		Name().
		// declares the values to be returned by the mocked function call.
		Return("Synthesized").
		// allows the expectation to be called 0 or more times
		AnyTimes()

	NewBalancerFactory(resolver, lb, Options{}) // note
	b, ok := balancerFactories.Load(cacheKey(t.Name(), "Synthesized", defaultOptions))
	test.Assert(t, ok)
	test.Assert(t, b != nil)
	bl, err := b.(*BalancerFactory).Get(context.Background(), nil)
	test.Assert(t, err == nil)
	test.Assert(t, bl.GetPicker() != nil)
	dump := Dump()
	dumpJson, err := json.Marshal(dump)
	test.Assert(t, err == nil)
	test.Assert(t, string(dumpJson) == `{"TestBuilder|Synthesized|{5s 15s}":{"mockRoute":[{"Address":"tcp://127.0.0.1:8888","Weight":10}]}}`)
}

func TestCacheKey(t *testing.T) {
	uniqueKey := cacheKey("hello", "world", Options{RefreshInterval: 15 * time.Second, ExpireInterval: 5 * time.Minute})
	test.Assert(t, uniqueKey == "hello|world|{15s 5m0s}")
}

func TestBalancerCache(t *testing.T) {
	count := 10
	inss := []discovery.Instance{}
	for i := 0; i < count; i++ {
		inss = append(inss, discovery.NewInstance("tcp", fmt.Sprint(i), 10, nil))
	}
	r := &discovery.SynthesizedResolver{
		TargetFunc: func(ctx context.Context, target rpcinfo.EndpointInfo) string {
			return target.ServiceName()
		},
		ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
			return discovery.Result{Cacheable: true, CacheKey: "svc", Instances: inss}, nil
		},
		NameFunc: func() string { return t.Name() },
	}
	lb := loadbalance.NewWeightedBalancer()
	for i := 0; i < count; i++ {
		blf := NewBalancerFactory(r, lb, Options{})
		info := rpcinfo.NewEndpointInfo("svc", "", nil, nil)
		b, err := blf.Get(context.Background(), info)
		test.Assert(t, err == nil)
		p := b.GetPicker()
		for a := 0; a < count; a++ {
			addr := p.Next(context.Background(), nil).Address().String()
			t.Logf("count: %d addr: %s\n", i, addr)
		}
	}
}

type mockRebalancer struct {
	rebalanceFunc func(ch discovery.Change)
	deleteFunc    func(ch discovery.Change)
}

// Rebalance implements the Rebalancer interface.
func (m *mockRebalancer) Rebalance(ch discovery.Change) {
	if m.rebalanceFunc != nil {
		m.rebalanceFunc(ch)
	}
}

// Delete implements the Rebalancer interface.
func (m *mockRebalancer) Delete(ch discovery.Change) {
	if m.deleteFunc != nil {
		m.deleteFunc(ch)
	}
}
