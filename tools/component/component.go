// Copyright © 2023 OpenIM open source community. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package component

import (
	"fmt"
	"github.com/OpenIMSDK/protocol/constant"
	"time"

	"github.com/OpenIMSDK/chat/pkg/common/config"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
)

func ComponentCheck(cfgPath string, hide bool) error {
	// If config.Config.Envs.Discovery != "k8s", perform Zookeeper checks
	// Note: Assuming the config is already loaded and available via config.Config

	if config.Config.Envs.Discovery != "k8s" {
		zkConn, err := checkNewZkClient(hide)
		if err != nil {
			errorPrint(fmt.Sprintf("%v.Please check if your openIM server has started", err.Error()), hide)
			return err
		}
		defer zkConn.Close()

		if err := checkGetCfg(zkConn, hide); err != nil {
			errorPrint(fmt.Sprintf("%v.Please check if your openIM server has started", err.Error()), hide)
			return err
		}
	}

	return nil
}

func errorPrint(s string, hide bool) {
	if !hide {
		fmt.Printf("\x1b[%dm%v\x1b[0m\n", 31, s)
	}
}

func successPrint(s string, hide bool) {
	if !hide {
		fmt.Printf("\x1b[%dm%v\x1b[0m\n", 32, s)
	}
}

func newZkClient() (*zk.Conn, error) {
	var c *zk.Conn
	c, _, err := zk.Connect(config.Config.Zookeeper.ZkAddr, time.Second, zk.WithLogger(log.NewZkLogger()))
	fmt.Println("zk addr=", config.Config.Zookeeper.ZkAddr)
	if err != nil {
		fmt.Println("zookeeper connect error:", err)
		return nil, errs.Wrap(err)
	} else {
		if config.Config.Zookeeper.Username != "" && config.Config.Zookeeper.Password != "" {
			if err := c.AddAuth("digest", []byte(config.Config.Zookeeper.Username+":"+config.Config.Zookeeper.Password)); err != nil {
				return nil, errs.Wrap(err)
			}
		}
	}
	return c, nil
}

func checkNewZkClient(hide bool) (*zk.Conn, error) {
	for i := 0; i < MaxConnectTimes; i++ {
		if i != 0 {
			time.Sleep(3 * time.Second)
		}
		zkConn, err := newZkClient()
		if err != nil {
			if zkConn != nil {
				zkConn.Close()
			}
			errorPrint(fmt.Sprintf("Starting Zookeeper failed: %v.Please make sure your Zookeeper service has started", err.Error()), hide)
			continue
		}
		successPrint(fmt.Sprint("zk starts successfully"), hide)
		return zkConn, nil
	}
	return nil, errs.Wrap(errors.New("Connecting to zk fails"))
}

func checkGetCfg(conn *zk.Conn, hide bool) error {
	for i := 0; i < MaxConnectTimes; i++ {
		if i != 0 {
			time.Sleep(3 * time.Second)
		}
		path := "/" + config.Config.Zookeeper.Schema + "/" + constant.OpenIMCommonConfigKey

		zkConfig, _, err := conn.Get(path)
		if err != nil {
			fmt.Println("path =", path, "zkConfig is:", zkConfig)
			errorPrint(fmt.Sprintf("! get zk config [%d] error: %v\n", i, err), hide)
			continue
		} else if len(zkConfig) == 0 {
			errorPrint(fmt.Sprintf("! get zk config [%d] data is empty\n", i), hide)
			continue
		}
		successPrint(fmt.Sprint("Chat get config successfully"), hide)
		return nil
	}
	return errs.Wrap(errors.New("Getting config from zk failed"))
}
