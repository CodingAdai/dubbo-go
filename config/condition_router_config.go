/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"encoding/base64"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

import (
	"github.com/apache/dubbo-go/cluster/directory"
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/logger"
)

var (
	mutex sync.Mutex
)

// ConditionRouterConfig Condition router config map to router config file
type ConditionRouterConfig struct {
	Priority   int      `yaml:"priority"`
	Force      bool     `yaml:"force" default:"false"`
	Conditions []string `yaml:"conditions"`
}

// Prefix Return router config prefix
func (*ConditionRouterConfig) Prefix() string {
	return constant.RouterConfigPrefix
}

//RouterInit Load config file to init router config
func RouterInit(confRouterFile string) error {
	routerConfig = &ConditionRouterConfig{}
	e := loadYmlConfig(confRouterFile, routerConfig)

	if e != nil {
		return e
	}

	logger.Debugf("router config{%#v}\n", routerConfig)
	directory.AddRouterURLSet(initRouterUrl(routerConfig))
	return nil
}

func initRouterUrl(routerConfig *ConditionRouterConfig) *common.URL {
	rule := parseCondition(routerConfig.Conditions)

	return common.NewURLWithOptions(
		common.WithProtocol(constant.ROUTE_PROTOCOL),
		common.WithIp(constant.ANYHOST_VALUE),
		common.WithParams(url.Values{}),
		common.WithParamsValue("force", strconv.FormatBool(routerConfig.Force)),
		common.WithParamsValue("priority", strconv.Itoa(routerConfig.Priority)),
		common.WithParamsValue(constant.RULE_KEY, base64.URLEncoding.EncodeToString([]byte(rule))),
		common.WithParamsValue("router", "condition"),
		common.WithParamsValue(constant.CATEGORY_KEY, constant.ROUTERS_CATEGORY))
}

func parseCondition(conditions []string) string {
	var (
		when string
		then string
	)
	for _, condition := range conditions {
		condition = strings.Trim(condition, " ")
		if strings.Contains(condition, "=>") {
			array := strings.SplitN(condition, "=>", 2)
			consumer := strings.Trim(array[0], " ")
			provider := strings.Trim(array[1], " ")
			if len(consumer) != 0 {
				if len(when) != 0 {
					when = strings.Join([]string{when, consumer}, " & ")
				} else {
					when = consumer
				}
			}
			if len(provider) != 0 {
				if len(then) != 0 {
					then = strings.Join([]string{then, provider}, " & ")
				} else {
					then = provider
				}
			}

		}

	}

	return strings.Join([]string{when, then}, " => ")
}
