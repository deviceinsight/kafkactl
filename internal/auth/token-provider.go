package auth

import (
	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/util"
	"github.com/deviceinsight/kafkactl/v5/pkg/plugins/auth"
)

var loadedPlugins = make(map[string]auth.AccessTokenProvider)

type pluginTokenProvider struct {
	pluginDelegate auth.AccessTokenProvider
}

func (p *pluginTokenProvider) Token() (*sarama.AccessToken, error) {
	token, err := p.pluginDelegate.Token()
	return &sarama.AccessToken{Token: token}, err
}

func LoadTokenProviderPlugin(pluginName string, options map[string]any, brokers []string) (sarama.AccessTokenProvider, error) {
	if pluginName == "generic" {
		return newGenericTokenProvider(options)
	}

	loadedPlugin, ok := loadedPlugins[pluginName]
	if !ok {
		var err error
		loadedPlugin, err = util.LoadPlugin(pluginName, auth.TokenProviderPluginSpec)
		if err != nil {
			return nil, err
		}

		if options == nil {
			options = make(map[string]any)
		}

		if err := loadedPlugin.Init(options, brokers); err != nil {
			return nil, err
		}
		loadedPlugins[pluginName] = loadedPlugin
	}

	return &pluginTokenProvider{loadedPlugin}, nil
}
