package parser

import (
	"io"
	"log"
	"os"

	"github.com/marcelriegr/draide/pkg/gittools"
	"github.com/marcelriegr/draide/pkg/ui"
	"github.com/valyala/fasttemplate"
)

// Template tbd
func Template(str string, contextDir string) string {
	// Get git repo information
	repoDetails, _ := gittools.GetRepoDetails(contextDir)

	// Parse environment variables
	str = Env(str)

	// Parse template variables
	t, err := fasttemplate.NewTemplate(str, "%", "%")
	if err != nil {
		log.Fatalf("unexpected error when parsing template: %s", err)
	}

	return t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		switch tag {
		case "COMMIT_HASH":
			if repoDetails == nil {
				ui.Error("Cannot resolve %%%s%% on a non git repository", tag)
				os.Exit(3)
			}
			return w.Write([]byte(repoDetails.CommitHash))
		case "BRANCH":
			if repoDetails == nil {
				ui.Error("Cannot resolve %%%s%% on a non git repository", tag)
				os.Exit(3)
			}
			return w.Write([]byte(repoDetails.Branch))
		default:
			ui.Error("Unrecognized tag template: %s", tag)
			os.Exit(2)
			panic(0)
		}
	})
}
