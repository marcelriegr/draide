package cmd

import (
	"github.com/marcelriegr/draide/internal/parser"
	"github.com/marcelriegr/draide/pkg/imgtools"
	"github.com/marcelriegr/draide/pkg/ui"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push an image",
	Long:  `Description tbd`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		repositoryFormat := viper.GetString("repository-format")
		templateVars := parser.GenerateTemplateVars(parser.GenerateTemplateVarsOptions{})
		tagTemplates := viper.GetStringSlice("tags")
		tags := parser.RepositoryName(repositoryFormat, tagTemplates, templateVars)

		if viper.GetBool("verbose") {
			ui.Log("Used configuration:")
			ui.Log("> repository name format: %s", repositoryFormat)
			ui.Log("> registry: %s", stringTernary(templateVars["REGISTRY"] == "", "<none>", templateVars["REGISTRY"]))
			ui.Log("> namespace: %s", stringTernary(templateVars["NAMESPACE"] == "", "<none>", templateVars["NAMESPACE"]))
			ui.Log("> base image name: %s", templateVars["IMAGE_NAME"])
			ui.Log("> tags:%s", stringTernary(len(tags) == 0, " <none>", ""))
			for _, v := range tags {
				ui.Log("  - %s", v)
			}
		}

		ui.Info("Pushing image...")
		if len(tags) == 0 {
			ui.ErrorAndExit(1, "Abort. No valid image tag found.")
		}
		for _, repository := range tags {
			imgtools.Push(repository, imgtools.PushOptions{
				Auth: imgtools.AuthConfig{
					Username: viper.GetString("username"),
					Password: viper.GetString("password"),
				},
			})
			ui.Success(" > %s pushed succefully", repository)
		}
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
