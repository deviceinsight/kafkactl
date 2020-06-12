package operations

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
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

func (operation *DocsOperation) GenerateDocs(rootCmd *cobra.Command, flags DocsFlags) error {

	if _, err := os.Stat(flags.Directory); os.IsNotExist(err) {
		if err := os.Mkdir(flags.Directory, os.ModePerm); err != nil {
			return errors.Wrap(err, "unable to create directory")
		}
	}

	switch flags.DocType {
	case "markdown", "mdown", "md":
		if err := doc.GenMarkdownTree(rootCmd, flags.Directory); err != nil {
			return errors.Wrap(err, "unable to generate markdown")
		}
		if flags.SinglePage {
			if err := generateSinglePage(flags); err != nil {
				return err
			}
		}
	case "man":
		manHdr := &doc.GenManHeader{Title: "KAFKACTL", Section: "1"}
		if err := doc.GenManTree(rootCmd, manHdr, flags.Directory); err != nil {
			return errors.Wrap(err, "unable to generate markdown")
		}
	default:
		return errors.Errorf("unknown doc type %q. Try 'markdown' or 'man'", flags.DocType)
	}
	return nil
}

func generateSinglePage(flags DocsFlags) error {

	files, err := ioutil.ReadDir(flags.Directory)
	if err != nil {
		return errors.Wrap(err, "unable to read files in directory")
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	singlePageMd := "kafkactl_docs.md"

	// Open a new file for writing only
	file, err := os.OpenFile(filepath.Join(flags.Directory, singlePageMd), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return errors.Wrap(err, "unable to open file")
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
			return errors.Wrap(err, "unable to open file")
		}

		if _, err = file.WriteString(content); err != nil {
			return errors.Wrap(err, "unable to write bytes")
		}

		if err := os.Remove(filename); err != nil {
			return errors.Wrap(err, "unable to remove file")
		}
	}

	output.Infof("File written: %s", filepath.Join(flags.Directory, singlePageMd))
	return nil
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
