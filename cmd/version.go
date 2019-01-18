package cmd

import (
	"fmt"
	"runtime"

	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

var version string
var buildTime string
var gitCommit string

type info struct {
	version   string
	buildTime string
	gitCommit string
	goVersion string
	compiler  string
	platform  string
}

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "print the version of kafkactl",
	Run: func(cmd *cobra.Command, args []string) {
		output.Infof("%#v", info{
			version:   version,
			buildTime: buildTime,
			gitCommit: gitCommit,
			goVersion: runtime.Version(),
			compiler:  runtime.Compiler,
			platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		})
	},
}

func init() {
	rootCmd.AddCommand(cmdVersion)
}
