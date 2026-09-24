package consume

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/global"
	"github.com/spf13/viper"
)

var protobufConfigOptions = []struct {
	name string
	want internal.ProtobufMarshalOptions
}{
	{"allowPartial", internal.ProtobufMarshalOptions{AllowPartial: true}},
	{"useProtoNames", internal.ProtobufMarshalOptions{UseProtoNames: true}},
	{"useEnumNumbers", internal.ProtobufMarshalOptions{UseEnumNumbers: true}},
	{"emitUnpopulated", internal.ProtobufMarshalOptions{EmitUnpopulated: true}},
	{"emitDefaultValues", internal.ProtobufMarshalOptions{EmitDefaultValues: true}},
}

func loadProtobufConfig(t *testing.T, contextName, configYAML string, env map[string]string) internal.ProtobufConfig {
	t.Helper()

	for _, option := range protobufConfigOptions {
		alias := "PROTOBUF_MARSHALOPTIONS_" + strings.ToUpper(option.name)
		for _, prefix := range []string{"", "CONTEXTS_DEFAULT_", "CONTEXTS_CONFIGURED_"} {
			t.Setenv(prefix+alias, "")
		}
	}
	for name, value := range env {
		t.Setenv(name, value)
	}

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(configFile, []byte(configYAML), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(global.ConfigEnvVariable, configFile)
	writableConfigFile := filepath.Join(dir, "current-context.yml")
	if err := os.WriteFile(writableConfigFile, []byte("current-context: "+contextName+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(global.WritableConfigEnvVariable, writableConfigFile)
	t.Setenv("CURRENT_CONTEXT", "")
	t.Cleanup(viper.Reset)

	config := global.NewConfig()
	config.Flags().ConfigFile = configFile
	config.Flags().Context = contextName
	if err := config.Init(); err != nil {
		t.Fatal(err)
	}
	context, err := internal.CreateClientContext()
	if err != nil {
		t.Fatal(err)
	}
	return context.Protobuf
}

func protobufOptionsYAML(contextName, options string) string {
	return fmt.Sprintf("contexts:\n  %s:\n    brokers: [localhost:9092]\n    protobuf:\n      marshalOptions:\n%s", contextName, options)
}

func TestConfiguredProtobufMarshalOptions(t *testing.T) {
	for _, contextName := range []string{"default", "configured"} {
		for _, option := range protobufConfigOptions {
			t.Run(contextName+"/"+option.name, func(t *testing.T) {
				yaml := protobufOptionsYAML(contextName, fmt.Sprintf("        %s: true\n", option.name))
				config := loadProtobufConfig(t, contextName, yaml, nil)
				if config.MarshalOptions != option.want {
					t.Fatalf("expected %+v, got %+v", option.want, config.MarshalOptions)
				}
			})
		}
	}

	for _, options := range []string{"", "        allowPartial: false\n        useProtoNames: false\n        useEnumNumbers: false\n        emitUnpopulated: false\n        emitDefaultValues: false\n"} {
		name := "unset"
		if options != "" {
			name = "explicit false"
		}
		t.Run(name, func(t *testing.T) {
			config := loadProtobufConfig(t, "default", protobufOptionsYAML("default", options), nil)
			if config.MarshalOptions != (internal.ProtobufMarshalOptions{}) {
				t.Fatalf("expected default options, got %+v", config.MarshalOptions)
			}
		})
	}
}

func TestProtobufMarshalOptionEnvironmentPrecedence(t *testing.T) {
	for _, option := range protobufConfigOptions {
		alias := "PROTOBUF_MARSHALOPTIONS_" + strings.ToUpper(option.name)
		for _, tc := range []struct {
			name        string
			contextName string
			yamlValue   bool
			env         map[string]string
			want        internal.ProtobufMarshalOptions
		}{
			{"alias overrides false", "default", false, map[string]string{alias: "true"}, option.want},
			{"alias overrides true", "default", true, map[string]string{alias: "false"}, internal.ProtobufMarshalOptions{}},
			{"prefixed overrides alias", "default", true, map[string]string{alias: "true", "CONTEXTS_DEFAULT_" + alias: "false"}, internal.ProtobufMarshalOptions{}},
			{"named context", "configured", false, map[string]string{"CONTEXTS_CONFIGURED_" + alias: "true"}, option.want},
			{"alias only affects default", "configured", false, map[string]string{alias: "true"}, internal.ProtobufMarshalOptions{}},
		} {
			t.Run(option.name+"/"+tc.name, func(t *testing.T) {
				yaml := protobufOptionsYAML(tc.contextName, fmt.Sprintf("        %s: %t\n", option.name, tc.yamlValue))
				config := loadProtobufConfig(t, tc.contextName, yaml, tc.env)
				if config.MarshalOptions != tc.want {
					t.Fatalf("expected %+v, got %+v", tc.want, config.MarshalOptions)
				}
			})
		}
	}
}
