package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认读取 ~/.codei18n/config.json 以及当前目录的 .codei18n/config.json)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "显示详细日志")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			cobra.CheckErr(err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		cobra.CheckErr(err)
	}

	configLoaded := false

	if loaded, err := loadConfigFile(filepath.Join(home, ".codei18n", "config.json"), false); err != nil {
		cobra.CheckErr(err)
	} else if loaded {
		configLoaded = true
	}

	if loaded, err := loadConfigFile(filepath.Join(".codei18n", "config.json"), configLoaded); err != nil {
		cobra.CheckErr(err)
	} else if loaded {
		configLoaded = true
	}

	if !configLoaded && verbose {
		fmt.Fprintln(os.Stderr, "未找到配置文件，使用内置默认配置")
	}
}

func loadConfigFile(path string, merge bool) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	viper.SetConfigFile(path)

	var err error
	if merge {
		err = viper.MergeInConfig()
	} else {
		err = viper.ReadInConfig()
	}
	if err != nil {
		return false, err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", path)
	}
	return true, nil
}

func main() {
	Execute()
}
