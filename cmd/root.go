package cmd

import (
	"fmt"

	"github.com/nexlycode/nexly/internal/config"
	"github.com/nexlycode/nexly/internal/tui"
	"github.com/spf13/cobra"
)

var (
	version   = "1.0.0"
	provider  string
	model     string
	temperature float64
	maxTokens int
)

	var rootCmd = &cobra.Command{
	Use:   "nexly",
	Short: "Nexly - AI Coding Assistant",
	Long:  `Nexly is a powerful CLI coding assistant that helps you write, edit, and understand code.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		
		if provider != "" {
			cfg.Provider = provider
			config.SaveConfig(&cfg)
			fmt.Printf("Provider set to: %s\n", provider)
			return
		}
		
		if model != "" {
			cfg.Model = model
			config.SaveConfig(&cfg)
			fmt.Printf("Model set to: %s\n", model)
			return
		}

		tui.Run(cfg)
	},
}

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage AI providers",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		fmt.Printf("Current provider: %s\n", cfg.Provider)
	},
}

var providerSetCmd = &cobra.Command{
	Use:   "set [provider]",
	Short: "Set the AI provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		cfg.Provider = args[0]
		config.SaveConfig(&cfg)
		fmt.Printf("Provider set to: %s\n", args[0])
	},
}

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage AI models",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		fmt.Printf("Current model: %s\n", cfg.Model)
	},
}

var modelSetCmd = &cobra.Command{
	Use:   "set [model]",
	Short: "Set the AI model",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		cfg.Model = args[0]
		config.SaveConfig(&cfg)
		fmt.Printf("Model set to: %s\n", args[0])
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		fmt.Printf("Provider: %s\n", cfg.Provider)
		fmt.Printf("Model: %s\n", cfg.Model)
		fmt.Printf("Temperature: %f\n", cfg.Temperature)
		fmt.Printf("MaxTokens: %d\n", cfg.MaxTokens)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Nexly version %s\n", version)
	},
}

func Execute() error {
	providerCmd.AddCommand(providerSetCmd)
	modelCmd.AddCommand(modelSetCmd)
	
	rootCmd.AddCommand(providerCmd)
	rootCmd.AddCommand(modelCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)

	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "", "Set AI provider")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "", "Set AI model")
	rootCmd.PersistentFlags().Float64VarP(&temperature, "temperature", "t", 0.7, "Set temperature")
	rootCmd.PersistentFlags().IntVarP(&maxTokens, "max-tokens", "M", 4096, "Set max tokens")

	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize()
}
