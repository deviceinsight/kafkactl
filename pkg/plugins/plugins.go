package plugins

import "github.com/hashicorp/go-plugin"

type PluginSpec[P plugin.Plugin, I any] struct {
	PluginImpl          P
	InterfaceIdentifier string
	Handshake           plugin.HandshakeConfig
}

func (s *PluginSpec[P, I]) GetMap() map[string]plugin.Plugin {
	return map[string]plugin.Plugin{
		s.InterfaceIdentifier: s.PluginImpl,
	}
}
