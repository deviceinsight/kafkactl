package cmd

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/cobra"
)

const docsDesc = `
Generate documentation files for kafkactl.
This command can generate documentation for Helm in the following formats:
- Markdown
- Man pages
`

var flags internal.DocsFlags

func newDocsCmd() *cobra.Command {

	var cmdDocs = &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation as markdown or man pages",
		Long:   docsDesc,
		Hidden: true,
		Run: func(cmd *cobra.Command, _ []string) {
			if err := (&internal.DocsOperation{}).GenerateDocs(cmd.Root(), flags); err != nil {
				output.Fail(err)
			}
		},
	}

	cmdDocs.Flags().StringVarP(&flags.Directory, "directory", "", "./", "directory to which documentation is written")
	cmdDocs.Flags().StringVarP(&flags.DocType, "type", "", "markdown", "the type of documentation to generate (markdown, man)")
	cmdDocs.Flags().BoolVarP(&flags.SinglePage, "single-page", "", false, "generate single page (markdown only)")

	return cmdDocs
}
