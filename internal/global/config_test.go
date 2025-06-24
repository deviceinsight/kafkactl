package global_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/global"
	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
	"github.com/spf13/viper"
)

func TestResolvePath(t *testing.T) {

	// switch working dir to git repo root
	rootDir, _ := testutil.GetRootDir()
	if err := os.Chdir(rootDir); err != nil {
		t.Fatalf("Error changing working directory to %s: %v", rootDir, err)
	}

	cfg := global.NewConfig()
	viper.SetConfigFile("cmd/config.yml")

	writable := viper.New()
	writable.SetConfigFile("pkg/config.yml")
	cfg.SetWritableConfig(writable)

	type testCases struct {
		description  string
		filename     string
		wantFilename string
		wantErr      string
	}

	for _, tc := range []testCases{
		{
			description:  "resolvable_relative_filename",
			filename:     "Makefile",
			wantFilename: "Makefile",
		},
		{
			description:  "resolvable_absolute_filename",
			filename:     filepath.Join(rootDir, "Makefile"),
			wantFilename: filepath.Join(rootDir, "Makefile"),
		},
		{
			description: "non_existent_absolute_filename",
			filename:    filepath.Join(rootDir, "NonExistent"),
			wantErr:     "NonExistent: no such file or directory",
		},
		{
			description: "non_resolvable_relative_filename",
			filename:    "NonExistent",
			wantErr:     "cannot find \"NonExistent\" in locations",
		},
		{
			description:  "filename_resolvable_relative_to_config",
			filename:     "root.go",
			wantFilename: "cmd/root.go",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {

			resolvedFilename, err := global.ResolvePath(tc.filename)

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("ResolvePath(%q): expected error, got nil", tc.filename)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("ResolvePath(%q): expected error %q, got %q", tc.filename, tc.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("ResolvePath(%q): unexpected error: %v", tc.filename, err)
				}
			}

			if resolvedFilename != tc.wantFilename {
				t.Fatalf("ResolvePath(%q): expected resolved filename %q, got %q", tc.filename, tc.wantFilename, resolvedFilename)
			}
		})
	}
}
