package cmd

import (
	"path/filepath"

	"github.com/marcelriegr/draide/internal/parser"
	"github.com/marcelriegr/draide/pkg/imgtools"
	"github.com/marcelriegr/draide/pkg/types"
	"github.com/marcelriegr/draide/pkg/ui"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	Use:   "build CONTEXT_DIR",
	Short: "Build an image",
	Long:  `Description tbd`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("dockerfile", cmd.PersistentFlags().Lookup("dockerfile"))
		viper.BindPFlag("nocache", cmd.PersistentFlags().Lookup("no-cache"))
		viper.BindPFlag("labels", cmd.PersistentFlags().Lookup("label"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		repositoryFormat := viper.GetString("repository-format")
		templateVars := parser.GenerateTemplateVars(parser.GenerateTemplateVarsOptions{})
		contextDir, _ := homedir.Expand(args[0])
		contextDir, err := filepath.Abs(contextDir)
		if err != nil {
			ui.Log(err.Error())
			ui.ErrorAndExit(3, "Failed parsing context directory path")
		}
		dockerfile := filepath.ToSlash(parser.Template(viper.GetString("dockerfile"), templateVars))
		noCache := viper.GetBool("nocache")

		push, err := cmd.Flags().GetBool("push")
		if err != nil {
			ui.ErrorAndExit(1, err.Error())
		}

		tagTemplates := viper.GetStringSlice("tags")
		tags := parser.RepositoryName(repositoryFormat, tagTemplates, templateVars)

		labelTemplates := viper.GetStringMapString("labels")
		labels := map[string]string{}
		for k, v := range labelTemplates {
			labels[k] = parser.Template(v, templateVars)
		}

		buildArgTemplates, err := cmd.Flags().GetStringToString("build-arg")
		if err != nil {
			ui.ErrorAndExit(1, err.Error())
		}
		if len(buildArgTemplates) == 0 {
			var buildArgsFromConfig []types.KeyValueConfig

			// unmarshal values into an interface as a workaround to enable case-sensitive data loading from config file
			// ref: https://github.com/spf13/viper/issues/373
			err = viper.UnmarshalKey("buildArgs", &buildArgsFromConfig)
			if err != nil {
				ui.Log(err.Error())
				ui.ErrorAndExit(1, "Failed parsing build arguments from configuration file")
			}

			for _, v := range buildArgsFromConfig {
				buildArgTemplates[v.Key] = v.Value
			}
		}
		buildArgs := map[string]string{}
		for k, v := range buildArgTemplates {
			buildArgs[k] = parser.Template(v, templateVars)
		}

		if viper.GetBool("verbose") {
			ui.Log("Used configuration:")
			ui.Log("> repository name format: %s", repositoryFormat)
			ui.Log("> registry: %s", stringTernary(templateVars["REGISTRY"] == "", "<none>", templateVars["REGISTRY"]))
			ui.Log("> namespace: %s", stringTernary(templateVars["NAMESPACE"] == "", "<none>", templateVars["NAMESPACE"]))
			ui.Log("> base image name: %s", templateVars["IMAGE_NAME"])
			ui.Log("> dockerfile: %s", dockerfile)
			ui.Log("> context: %s", contextDir)
			ui.Log("> no-cache: %v", noCache)
			ui.Log("> labels:%s", stringTernary(len(labels) == 0, " <none>", ""))
			for k, v := range labels {
				ui.Log("  - %s: %s", k, v)
			}
			ui.Log("> build args:%s", stringTernary(len(buildArgs) == 0, " <none>", ""))
			for k, v := range buildArgs {
				ui.Log("  - %s: %s", k, v)
			}
			ui.Log("> tags:%s", stringTernary(len(tags) == 0, " <none>", ""))
			for _, v := range tags {
				ui.Log("  - %s", v)
			}
		}

		ui.Info("Building image...")
		if len(tags) == 0 {
			ui.ErrorAndExit(1, "Abort. No valid image tag found.")
		}
		imgtools.Build(contextDir, imgtools.BuildOptions{
			Dockerfile: dockerfile,
			BuildArgs:  buildArgs,
			Tags:       tags,
			Labels:     labels,
			NoCache:    noCache,
		})

		for _, repository := range tags {
			ui.Success(" > %s built succefully", repository)
		}

		if push {
			ui.Info("Pushing image...")
			for _, repository := range tags {
				imgtools.Push(repository, imgtools.PushOptions{
					Auth: imgtools.AuthConfig{
						Username: viper.GetString("username"),
						Password: viper.GetString("password"),
					},
				})
				ui.Success(" > %s pushed succefully", repository)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.PersistentFlags().StringP("dockerfile", "f", "Dockerfile", "Path to Dockerfile relative to CONTEXT_DIR. Value may contain template variable.")
	buildCmd.PersistentFlags().StringToString("label", map[string]string{}, "Image label. Value may contain template variable.")
	buildCmd.PersistentFlags().StringToString("build-arg", map[string]string{}, "Build argument. Value may contain template variable.")
	buildCmd.PersistentFlags().Bool("no-cache", false, "Set build noCache option")
	buildCmd.PersistentFlags().Bool("push", false, "Push image after building")
}

func stringTernary(condition bool, trueValue string, falseValue string) string {
	if condition {
		return trueValue
	}
	return falseValue
}
