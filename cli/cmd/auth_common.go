package cmd

import (
	"fmt"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
)

// cognitoConfigs 硬编码的 Cognito 配置，按环境区分
var cognitoConfigs = map[string]*internal.CognitoConfig{
	"dev": {
		UserPoolID:   "us-west-2_8D8Pp5UcN",
		ClientID:     "7rjsdgrfm5iuvvia7m2mu8e67a",
		Domain:       "lexi2-dev.auth.us-west-2.amazoncognito.com",
		Region:       "us-west-2",
		CallbackPort: 8888,
	},
	"preprod": {
		UserPoolID:   "us-west-2_8D8Pp5UcN", // TODO: 更新为 preprod 的配置
		ClientID:     "7rjsdgrfm5iuvvia7m2mu8e67a",
		Domain:       "lexi2-dev.auth.us-west-2.amazoncognito.com",
		Region:       "us-west-2",
		CallbackPort: 8888,
	},
	"prod": {
		UserPoolID:   "us-west-2_8D8Pp5UcN", // TODO: 更新为 prod 的配置
		ClientID:     "7rjsdgrfm5iuvvia7m2mu8e67a",
		Domain:       "lexi2-dev.auth.us-west-2.amazoncognito.com",
		Region:       "us-west-2",
		CallbackPort: 8888,
	},
}

// loadCognitoConfig 从硬编码配置加载 Cognito 配置。
func loadCognitoConfig(env string) (*internal.CognitoConfig, error) {
	config, ok := cognitoConfigs[env]
	if !ok {
		return nil, fmt.Errorf("no Cognito config for environment '%s'", env)
	}
	return config, nil
}
