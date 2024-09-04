package cmd

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/deviceinsight/kafkactl/v5/internal/global"

	"github.com/deviceinsight/kafkactl/v5/cmd/alter"
	"github.com/deviceinsight/kafkactl/v5/cmd/attach"
	"github.com/deviceinsight/kafkactl/v5/cmd/clone"
	"github.com/deviceinsight/kafkactl/v5/cmd/config"
	"github.com/deviceinsight/kafkactl/v5/cmd/consume"
	"github.com/deviceinsight/kafkactl/v5/cmd/create"
	"github.com/deviceinsight/kafkactl/v5/cmd/deletion"
	"github.com/deviceinsight/kafkactl/v5/cmd/describe"
	"github.com/deviceinsight/kafkactl/v5/cmd/get"
	"github.com/deviceinsight/kafkactl/v5/cmd/produce"
	"github.com/deviceinsight/kafkactl/v5/cmd/reset"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/cobra"
)

func NewKafkactlCommand(streams output.IOStreams) *cobra.Command {

	var rootCmd = &cobra.Command{
		Use:   "kafkactl",
		Short: "command-line interface for Apache Kafka",
		Long:  `A command-line interface the simplifies interaction with Kafka.`,
	}

	globalConfig := global.NewConfig()

	cobra.OnInitialize(globalConfig.Init)

	rootCmd.AddCommand(config.NewConfigCmd())
	rootCmd.AddCommand(consume.NewConsumeCmd())
	rootCmd.AddCommand(create.NewCreateCmd())
	rootCmd.AddCommand(alter.NewAlterCmd())
	rootCmd.AddCommand(deletion.NewDeleteCmd())
	rootCmd.AddCommand(describe.NewDescribeCmd())
	rootCmd.AddCommand(get.NewGetCmd())
	rootCmd.AddCommand(produce.NewProduceCmd())
	rootCmd.AddCommand(reset.NewResetCmd())
	rootCmd.AddCommand(attach.NewAttachCmd())
	rootCmd.AddCommand(clone.NewCloneCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newDocsCmd())

	globalFlags := globalConfig.Flags()

	// use upper-case letters for shorthand params to avoid conflicts with local flags
	rootCmd.PersistentFlags().StringVarP(&globalFlags.ConfigFile, "config-file", "C", "",
		fmt.Sprintf("config file. default locations: %v", globalConfig.DefaultPaths()))
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Verbose, "verbose", "V", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&globalFlags.Context, "context", "", "The name of the context to use")

	err := rootCmd.RegisterFlagCompletionFunc("context", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var contexts []string
		for k := range viper.GetStringMap("contexts") {
			contexts = append(contexts, k)
		}
		return contexts, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		panic(err)
	}

	k8s.KafkaCtlVersion = Version

	output.IoStreams = streams
	rootCmd.SetOut(streams.Out)
	rootCmd.SetErr(streams.ErrOut)
	return rootCmd
}
