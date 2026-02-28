package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Provider    string            `json:"provider"`
	Model       string            `json:"model"`
	Temperature float64          `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
	APIKeys     map[string]string `json:"api_keys"`
	History     []Message         `json:"history"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var defaultConfig = Config{
	Provider:    "openai",
	Model:       "gpt-4",
	Temperature: 0.7,
	MaxTokens:   4096,
	APIKeys:     make(map[string]string),
	History:     []Message{},
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".nexly", "config.json")
}

func ensureConfigDir() error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".nexly")
	return os.MkdirAll(dir, 0700)
}

func LoadConfig() Config {
	if err := ensureConfigDir(); err != nil {
		return defaultConfig
	}

	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		cfg := defaultConfig
		cfg.APIKeys = make(map[string]string)
		cfg.History = []Message{}
		return cfg
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig
	}

	if cfg.APIKeys == nil {
		cfg.APIKeys = make(map[string]string)
	}
	if cfg.History == nil {
		cfg.History = []Message{}
	}

	return cfg
}

func SaveConfig(cfg *Config) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath(), data, 0600)
}

func GetAPIKey(provider string) string {
	cfg := LoadConfig()
	return cfg.APIKeys[provider]
}

func SetAPIKey(provider, key string) error {
	cfg := LoadConfig()
	cfg.APIKeys[provider] = key
	return SaveConfig(&cfg)
}

func AddMessage(role, content string) error {
	cfg := LoadConfig()
	cfg.History = append(cfg.History, Message{
		Role:    role,
		Content: content,
	})
	
	if len(cfg.History) > 100 {
		cfg.History = cfg.History[len(cfg.History)-100:]
	}
	
	return SaveConfig(&cfg)
}

func ClearHistory() error {
	cfg := LoadConfig()
	cfg.History = []Message{}
	return SaveConfig(&cfg)
}

func GetModels(provider string) []string {
	switch provider {
	case "openai":
		return []string{
			"gpt-4-turbo",
			"gpt-4",
			"gpt-4o",
			"gpt-4o-mini",
			"gpt-3.5-turbo",
			"o1",
			"o1-mini",
			"o1-preview",
		}
	case "anthropic":
		return []string{
			"claude-3-5-sonnet-20241022",
			"claude-3-5-sonnet-20240620",
			"claude-3-opus-20240229",
			"claude-3-haiku-20240307",
		}
	case "google":
		return []string{
			"gemini-2.0-flash",
			"gemini-1.5-pro",
			"gemini-1.5-flash",
			"gemini-1.0-pro",
		}
	case "openrouter":
		return []string{
			"openai/gpt-4",
			"openai/gpt-4o",
			"anthropic/claude-3.5-sonnet",
			"google/gemini-pro-1.5",
			"meta-llama/llama-3.1-70b-instruct",
		}
	case "nvidia":
		return []string{
			"nvidia/llama-3.1-nemotron-70b-instruct",
			"nvidia/mixtral-8x7b-instruct-v0.1",
			"nvidia/mistral-7b-instruct-v0.2",
		}
	default:
		return []string{"gpt-4"}
	}
}

func GetProviders() []string {
	return []string{"openai", "anthropic", "google", "openrouter", "nvidia"}
}
