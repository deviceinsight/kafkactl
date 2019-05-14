package operations

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type DocsFlags struct {
	Directory  string
	DocType    string
	SinglePage bool
}

type DocsOperation struct {
}

func (operation *DocsOperation) GenerateDocs(rootCmd *cobra.Command, flags DocsFlags) {

	if _, err := os.Stat(flags.Directory); os.IsNotExist(err) {
		if err := os.Mkdir(flags.Directory, os.ModePerm); err != nil {
			output.Failf("unable to create directory: %v", err)
		}
	}

	switch flags.DocType {
	case "markdown", "mdown", "md":
		if err := doc.GenMarkdownTree(rootCmd, flags.Directory); err != nil {
			output.Failf("unable to generate markdown: %v", err)
		}
		if flags.SinglePage {
			generateSinglePage(flags)
		}
	case "man":
		manHdr := &doc.GenManHeader{Title: "KAFKACTL", Section: "1"}
		if err := doc.GenManTree(rootCmd, manHdr, flags.Directory); err != nil {
			output.Failf("unable to generate markdown: %v", err)
		}
	default:
		output.Failf("unknown doc type %q. Try 'markdown' or 'man'", flags.DocType)
	}
}

func generateSinglePage(flags DocsFlags) {

	files, err := ioutil.ReadDir(flags.Directory)
	if err != nil {
		output.Failf("unable to read files in directory: %v", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	singlePageMd := "kafkactl_docs.md"

	// Open a new file for writing only
	file, err := os.OpenFile(filepath.Join(flags.Directory, singlePageMd), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		output.Failf("unable to open file: %v", err)
	}
	defer file.Close()

	for _, f := range files {

		if f.Name() == singlePageMd || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}

		filename := filepath.Join(flags.Directory, f.Name())
		bytes, err := ioutil.ReadFile(filename)

		content := string(bytes)

		content = adjustChapters(f.Name(), content)
		content = removeFooter(content)

		if err != nil {
			output.Failf("unable to open file: %v", err)
		}

		if _, err = file.WriteString(content); err != nil {
			output.Failf("unable to write bytes: %v", err)
		}

		if err := os.Remove(filename); err != nil {
			output.Failf("unable to remove file: %v", err)
		}
	}

	output.Infof("File written: %s", filepath.Join(flags.Directory, singlePageMd))
}

func adjustChapters(filename string, content string) string {
	separatorRegex := regexp.MustCompile("_")
	matches := separatorRegex.FindAllStringIndex(filename, -1)

	startLayer := len(matches)

	chapterRegex := regexp.MustCompile("(?m)^(#+.*)$")

	return chapterRegex.ReplaceAllString(content, strings.Repeat("#", startLayer)+"$1")
}

func removeFooter(content string) string {
	footerRegex := regexp.MustCompile("(?m)^#+ Auto generated.*$")
	return footerRegex.ReplaceAllLiteralString(content, "")
}
