package validation

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	BashCompAtLeastOneRequiredFlag = "cobra_annotation_bash_completion_at_least_one_required_flag"
)

func MarkFlagAtLeastOneRequired(flags *pflag.FlagSet, name string) error {
	return flags.SetAnnotation(name, BashCompAtLeastOneRequiredFlag, []string{"true"})
}

func ValidateAtLeastOneRequiredFlag(cmd *cobra.Command) error {
	requiredError := true
	atLeastRequiredFlags := []string{}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		atLeastOneRequiredAnnotation := flag.Annotations[BashCompAtLeastOneRequiredFlag]
		if len(atLeastOneRequiredAnnotation) == 0 {
			return
		}

		flagRequired := atLeastOneRequiredAnnotation[0] == "true"

		if flagRequired {
			atLeastRequiredFlags = append(atLeastRequiredFlags, flag.Name)
		}

		if flag.Changed {
			requiredError = false
		}
	})

	if requiredError {
		return errors.New("At least one of the following flags must be set: " + strings.Join(atLeastRequiredFlags[:], ", "))
	}

	return nil
}
