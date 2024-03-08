package cmd

import (
	"fmt"
	"runtime"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/cobra"
)

var Version string
var BuildTime string
var GitCommit string

type info struct {
	version   string
	buildTime string
	gitCommit string
	goVersion string
	compiler  string
	platform  string
}

func newVersionCmd() *cobra.Command {
	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "print the version of kafkactl",
		Run: func(_ *cobra.Command, _ []string) {
			output.Infof("%#v", info{
				version:   Version,
				buildTime: BuildTime,
				gitCommit: GitCommit,
				goVersion: runtime.Version(),
				compiler:  runtime.Compiler,
				platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			})
		},
	}
	return cmdVersion
}
