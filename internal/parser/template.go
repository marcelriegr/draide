package parser

import (
	"io"

	"github.com/marcelriegr/draide/pkg/gittools"
	"github.com/marcelriegr/draide/pkg/ui"

	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"
)

// TemplateVars tbd
type TemplateVars map[string]string

// GenerateTemplateVarsOptions tbd
type GenerateTemplateVarsOptions struct {
	contextDir string
}

// GenerateTemplateVars tbd
func GenerateTemplateVars(opts GenerateTemplateVarsOptions) TemplateVars {
	vars := map[string]string{
		"IMAGE_NAME": viper.GetString("imagename"),
		"REGISTRY":   viper.GetString("registry"),
		"NAMESPACE":  viper.GetString("namespace"),
	}

	if opts.contextDir == "" {
		opts.contextDir = "."
	}

	repoDetails, _ := gittools.GetRepoDetails(opts.contextDir)
	if repoDetails != nil {
		vars["BRANCH"] = repoDetails.Branch
		vars["COMMIT_HASH"] = repoDetails.CommitHash
	}

	return vars
}

// Template tbd
func Template(template string, templateVars TemplateVars) string {
	// Interpolate environment variables
	template = Env(template)

	// Parse template
	t, err := fasttemplate.NewTemplate(template, "%", "%")
	if err != nil {
		ui.Log(err.Error())
		ui.ErrorAndExit(1, "Failed parsing template %s", template)
	}

	// Interpolate template variables
	return t.ExecuteFuncString(func(w io.Writer, templateVar string) (int, error) {
		val, validKey := templateVars[templateVar]

		if !validKey {
			ui.ErrorAndExit(1, "Unrecognized template variable: %s", templateVar)
		}

		if val == "" {
			switch templateVar {
			case "BRANCH":
			case "COMMIT_HASH":
				ui.ErrorAndExit(1, "Cannot resolve %%%s%% on a non git repository", templateVar)
			}
			ui.ErrorAndExit(1, "Cannot resolve template variable %s", templateVar)
		}

		return w.Write([]byte(val))
	})
}
