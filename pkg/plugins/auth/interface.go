package auth

import (
	"github.com/deviceinsight/kafkactl/v5/pkg/plugins"
	"github.com/hashicorp/go-plugin"
)

type AccessTokenProvider interface {
	Token() (string, error)
	Init(options map[string]any, brokers []string) error
}

var TokenProviderPluginSpec = plugins.PluginSpec[*TokenProviderPlugin, AccessTokenProvider]{
	PluginImpl: &TokenProviderPlugin{},
	Handshake: plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "KAFKACTL_PLUGIN",
		MagicCookieValue: "TOKEN_PROVIDER_PLUGIN",
	},
}
