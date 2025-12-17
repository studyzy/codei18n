package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	version = "dev" // Override at build time via -ldflags "-X main.version=<version>"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "codei18n",
	Short: "CodeI18n 核心 CLI 工具",
	Long: `CodeI18n 是一个代码注释国际化基础设施工具。
它支持源码注释扫描、ID 生成、多语言映射管理以及自动翻译功能。
主要用于将源码中的英文注释与本地语言注释进行互转。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Version = version
	rootCmd.SetVersionTemplate("CodeI18n CLI 版本: {{.Version}}\n")

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认是 $HOME/.codei18n.json 或项目根目录 .codei18n/config.json)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "显示详细日志")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			cobra.CheckErr(err)
		}

		// Search config in home directory with name ".codei18n" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".codei18n")
		viper.SetConfigType("json")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

func main() {
	Execute()
}
