package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcelriegr/draide/pkg/ui"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var preset string
var imageName string = ""

var rootCmd = &cobra.Command{
	Use:   "draide",
	Short: "Dr Aide - Your personal Docker aide",
	Long:  `Utility tools to build and publish Docker image`,
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (defaults to ./.draide.yaml or $HOME/.draide.yaml)")

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

	rootCmd.PersistentFlags().StringSliceP("tag", "t", []string{}, "Image tag. Value may contains template variable.")
	viper.BindPFlag("tags", rootCmd.PersistentFlags().Lookup("tag"))
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
			viper.MergeConfigMap(viper.Sub(presetsKey).AllSettings())
		}
	case viper.ConfigFileNotFoundError:
		ui.Log("Proceed without configuration file")
	case viper.ConfigParseError:
		ui.ErrorAndExit(1, "Error while parsing configuration file")
	default:
		ui.ErrorAndExit(1, err.Error())
	}

}
