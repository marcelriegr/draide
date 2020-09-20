package cmd

import (
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	"github.com/marcelriegr/draide/internal/parser"
	"github.com/marcelriegr/draide/pkg/imgtools"
	"github.com/marcelriegr/draide/pkg/ui"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	Use:   "build CONTEXT_DIR",
	Short: "Build an image",
	Long: `Description tbd
	
Available template variables:
 $<ENV_VAR>			Environment variable
 #<ENV_VAR>			Alias for $<ENV_VAR> as the syntax may be evaluated by the system and thus requires escaping to be passed correctly
 %BRANCH%			Git branch name of current directory
 %COMMIT_HASH%			Git commit hash of current directory
`,
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("dockerfile", cmd.PersistentFlags().Lookup("dockerfile"))
		viper.BindPFlag("labels", cmd.PersistentFlags().Lookup("label"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		contextDir, _ := homedir.Expand(args[0])
		contextDir, err := filepath.Abs(contextDir)
		if err != nil {
			ui.Log(err.Error())
			ui.ErrorAndExit(3, "Failed parsing context directory path")
		}
		imageName := viper.GetString("imagename")
		dockerfile := filepath.ToSlash(parser.Template(viper.GetString("dockerfile"), contextDir))

		noCache, err := cmd.Flags().GetBool("no-cache")
		if err != nil {
			ui.ErrorAndExit(1, err.Error())
		}

		tagsUnparsed := viper.GetStringSlice("tags")
		tags := make([]string, len(tagsUnparsed))
		for i, tag := range tagsUnparsed {
			tags[i] = parser.Template(tag, contextDir)
		}

		labelsUnparsed := viper.GetStringMapString("labels")
		labels := map[string]string{}
		for k, v := range labelsUnparsed {
			labels[k] = parser.Template(v, contextDir)
		}

		buildArgsUnparsed, err := cmd.Flags().GetStringToString("build-arg")
		if err != nil {
			ui.ErrorAndExit(1, err.Error())
		}
		if len(buildArgsUnparsed) == 0 {
			var buildArgsFromConfig []struct {
				Key   string
				Value string
			}
			// unmarshal values into an interface as a workaround to enable case-sensitive data loading from config file
			// ref: https://github.com/spf13/viper/issues/373
			viper.UnmarshalKey("buildArgs", &buildArgsFromConfig)
			for _, v := range buildArgsFromConfig {
				buildArgsUnparsed[v.Key] = v.Value
			}
		}
		buildArgs := map[string]string{}
		for k, v := range buildArgsUnparsed {
			buildArgs[k] = parser.Template(v, contextDir)
		}

		ui.Info("Building image...")
		if viper.GetBool("verbose") {
			ui.Log("> image name: %s", imageName)
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
				ui.Log("  - %s:%s", imageName, v)
			}
		}

		imgtools.Build(imageName, contextDir, imgtools.BuildOptions{
			Dockerfile: dockerfile,
			BuildArgs:  buildArgs,
			Tags:       tags,
			Labels:     labels,
			NoCache:    noCache,
		})
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.PersistentFlags().StringP("dockerfile", "f", "Dockerfile", "Path to Dockerfile. Value may contains template variable.")
	buildCmd.PersistentFlags().StringToString("label", map[string]string{}, "Image label. Value may contains template variable.")
	buildCmd.PersistentFlags().StringToString("build-arg", map[string]string{}, "Build argument. Value may contains template variable.")
	buildCmd.PersistentFlags().Bool("no-cache", false, "Set build noCache option")
}

func stringTernary(condition bool, trueValue string, falseValue string) string {
	if condition {
		return trueValue
	}
	return falseValue
}
