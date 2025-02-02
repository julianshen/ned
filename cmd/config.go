package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type Config struct {
	Values map[string]string `toml:"values"`
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage ned configuration settings`,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration value. For example: ned config set ANTHROPIC_API_KEY your-key-here`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("accepts 2 arg(s), received %d", len(args))
		}
		key := args[0]
		value := args[1]

		config, err := loadConfig()
		if err != nil {
			return err
		}

		if config.Values == nil {
			config.Values = make(map[string]string)
		}

		if existingValue, exists := config.Values[key]; exists {
			fmt.Printf("Key '%s' already exists with value '%s'. Do you want to override it? (y/N): ", key, existingValue)
			var response string
			fmt.Scanln(&response)
			if !strings.EqualFold(response, "y") {
				return nil
			}
		}

		config.Values[key] = value
		return saveConfig(config)
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show configuration",
	Long:  `Show all configuration values`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		if len(config.Values) == 0 {
			fmt.Println("No configuration values set")
			return nil
		}

		for key, value := range config.Values {
			fmt.Printf("%s: %s\n", key, value)
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "ned")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.toml"), nil
}

func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	config := &Config{}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return nil, err
	}

	return config, nil
}

func saveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(config)
}
