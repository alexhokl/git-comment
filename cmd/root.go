package cmd

import (
	"github.com/alexhokl/helper/cli"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:          "git-comment",
	Short:        "Generate git comment using models from Ollama",
	SilenceUsage: true,
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-comment.yml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	cli.ConfigureViper(cfgFile, "git-comment", false, "")
}
