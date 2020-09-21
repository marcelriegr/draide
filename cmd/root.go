package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcelriegr/draide/pkg/types"
	"github.com/marcelriegr/draide/pkg/ui"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var preset string
var imageName string = ""

var rootCmd = &cobra.Command{
	Use:   "draide",
	Short: "Dr Aide - Your personal Docker aide",
	Long: `Utility tools to build and publish Docker image

Available template variables:
	$<ENV_VAR>			Environment variable
	#<ENV_VAR>			Alias for $<ENV_VAR> as the syntax may get evaluated by the system and thus requires escaping to be passed correctly
	%REGISTRY%			Registry (see --registry flag)
	%NAMESPACE%			Namespace (see --namespace flag)
	%IMAGE_NAME%			Image name (see --name flag)
	%BRANCH%			Git branch name of current directory
	%COMMIT_HASH%			Git commit hash of current directory
`,
}

// Execute Cobra
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Logging verbosity")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ./.draide.yaml or $HOME/.draide.yaml)")

	rootCmd.PersistentFlags().StringVarP(&preset, "preset", "p", "", "Use presets")
	viper.BindPFlag("preset", rootCmd.PersistentFlags().Lookup("preset"))

	rootCmd.PersistentFlags().StringVarP(&imageName, "name", "n", "", "Image name. (default <directory-name>)")
	viper.BindPFlag("imageName", rootCmd.PersistentFlags().Lookup("name"))
	cwd, err := os.Getwd()
	if err != nil {
		ui.ErrorAndExit(1, err.Error())
	}
	if imageName == "" {
		imageName = filepath.Base(cwd)
	}

	rootCmd.PersistentFlags().StringSliceP("tag", "t", []string{}, "Image tag. Value may contain template variable.")
	viper.BindPFlag("tags", rootCmd.PersistentFlags().Lookup("tag"))
	viper.SetDefault("tags", []string{"latest"})

	rootCmd.PersistentFlags().StringP("registry", "r", "", "Container registry, such as: k8s.gcr.io")
	viper.BindPFlag("registry", rootCmd.PersistentFlags().Lookup("registry"))

	rootCmd.PersistentFlags().String("namespace", "", "Repository namespace")
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))

	rootCmd.PersistentFlags().String("repository-format", "%REGISTRY%/%NAMESPACE%/%IMAGE_NAME%", "Format to construct repository name. Value may contain template variable.")
	viper.BindPFlag("repository-format", rootCmd.PersistentFlags().Lookup("repository-format"))

	rootCmd.PersistentFlags().String("username", "", "Username for pushing image into registry")
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))

	rootCmd.PersistentFlags().String("password", "", "Password for pushing image into registry")
	rootCmd.PersistentFlags().Bool("password-stdin", false, "Password for pushing image into registry via stdin. Password supplied via stdin will take precedence over the --password flag.")
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.SetConfigName(".draide")
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
	}

	viper.SetEnvPrefix("draide")
	viper.AutomaticEnv()

	switch err := viper.ReadInConfig(); err.(type) {
	case nil:
		ui.Info("Using configuration file: %s", viper.ConfigFileUsed())

		if preset != "" {
			presetsKey := "presets." + preset
			if !viper.IsSet(presetsKey) {
				ui.ErrorAndExit(1, "Unable to find preset configuration for: %s", preset)
			}

			presetSettings := viper.Sub(presetsKey)
			err = viper.MergeConfigMap(viper.Sub(presetsKey).AllSettings())
			if err != nil {
				ui.Log(err.Error())
				ui.ErrorAndExit(1, "Failed parsing preset configuration")
			}

			if presetSettings.IsSet("extraBuildArgs") {
				var buildArgs, buildArgsOfPreset []types.KeyValueConfig

				// unmarshal values into an interface as a workaround to enable case-sensitive data loading from config file
				// ref: https://github.com/spf13/viper/issues/373
				err = viper.UnmarshalKey("buildArgs", &buildArgs)
				if err != nil {
					ui.Log(err.Error())
					ui.ErrorAndExit(1, "Failed parsing build arguments from configuration file")
				}
				err = presetSettings.UnmarshalKey("extraBuildArgs", &buildArgsOfPreset)
				if err != nil {
					ui.Log(err.Error())
					ui.ErrorAndExit(1, "Failed parsing preset's extra build arguments from configuration file")
				}
				viper.Set("buildArgs", append(buildArgs, buildArgsOfPreset...))
			}

			if presetSettings.IsSet("extraTags") {
				tags := viper.GetStringSlice("tags")
				extraTags := viper.GetStringSlice("extraTags")
				viper.Set("tags", append(tags, extraTags...))
			}
		}
	case viper.ConfigFileNotFoundError:
		ui.Log("Proceed without configuration file")
	case viper.ConfigParseError:
		ui.ErrorAndExit(1, "Error while parsing configuration file")
	default:
		ui.ErrorAndExit(1, err.Error())
	}

	initCredentials()
}

func initCredentials() {
	// read credentials from stdin
	passwordStdIn, err := rootCmd.PersistentFlags().GetBool("password-stdin")
	if err != nil {
		ui.ErrorAndExit(1, err.Error())
	}

	if passwordStdIn {
		fi, err := os.Stdin.Stat()
		if err != nil {
			ui.ErrorAndExit(1, err.Error())
		} else if fi.Mode()&os.ModeNamedPipe == 0 {
			ui.ErrorAndExit(1, "No password from stdin received")
		}

		scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
		scanner.Scan()
		password := scanner.Text()

		if password == "" {
			ui.ErrorAndExit(1, "Found empty string as password from stdin")
		}

		viper.Set("password", password)
		ui.Log("Using password from stdin")
	}

	// check credentials completeness
	username := viper.GetString("username")
	password := viper.GetString("password")
	if (username == "" && password != "") || (username != "" && password == "") {
		ui.ErrorAndExit(1, "Incomplete credentials. Username and password information must be provided.")
	}

	if ui.IsVerbose() && username != "" && password != "" {
		ui.Log("Credentials:")
		ui.Log(" > username: %s", username)
		ui.Log(" > password: ******")
	}
}
