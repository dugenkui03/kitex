// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"flag"
	"os"
	"strings"

	"github.com/cloudwego/kitex"
	kargs "github.com/cloudwego/kitex/tool/cmd/kitex/args"
	"github.com/cloudwego/kitex/tool/internal_pkg/pluginmode/protoc"
	"github.com/cloudwego/kitex/tool/internal_pkg/pluginmode/thriftgo"
)

// 初始化的时候就添加 -version参数 相关的内容
var args kargs.Arguments

func init() {
	var queryVersion bool
	args.AddExtraFlag(&kargs.ExtraFlag{
		// 获取 kitex 的版本类型
		Apply: func(f *flag.FlagSet) {
			f.BoolVar(&queryVersion, "version", false,
				"Show the version of kitex")
		},
		Check: func(a *kargs.Arguments) {
			if queryVersion {
				println(a.Version)
				os.Exit(0)
			}
		},
	})
}

func main() {
	// KITEX_PLUGIN_MODE kitex 查询运行模式
	mode := os.Getenv(kargs.EnvPluginMode)
	if len(os.Args) <= 1 && mode != "" {
		// run as a plugin
		// note thrift模式 和 proto模式
		switch mode {
		case thriftgo.PluginName:
			os.Exit(thriftgo.Run())
		case protoc.PluginName:
			os.Exit(protoc.Run())
		}
	}

	// run as kitex
	args.ParseArgs(kitex.Version)

	out := new(bytes.Buffer)
	cmd := args.BuildCmd(out)
	err := cmd.Run()
	if err != nil {
		if args.Use != "" {
			out := strings.TrimSpace(out.String())
			if strings.HasSuffix(out, thriftgo.TheUseOptionMessage) {
				os.Exit(0)
			}
		}
		os.Exit(1)
	}
}
