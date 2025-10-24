package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Database   DatabaseConfig   `yaml:"database"`
	Queries    QueriesConfig    `yaml:"queries"`
	Data       DataConfig       `yaml:"data"`
	Logging    LoggingConfig    `yaml:"logging"`
	OpenAI     OpenAIConfig     `yaml:"openai"`
	Prompts    PromptsConfig    `yaml:"prompts"`
	Batch      BatchConfig      `yaml:"batch"`
	RateLimit  RateLimitConfig  `yaml:"rate_limit"`
	Retry      RetryConfig      `yaml:"retry"`
	Formatting FormattingConfig `yaml:"formatting"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	DBName         string `yaml:"dbname"`
	SSLMode        string `yaml:"sslmode"`
	MaxIdleConns   int    `yaml:"max_idle_conns"`
	MaxOpenConns   int    `yaml:"max_open_conns"`
	MaxLifetimeMin int    `yaml:"max_lifetime_minutes"`
}

// QueriesConfig holds SQL queries
type QueriesConfig struct {
	ProfilesKid         string `yaml:"profiles_kid"`
	Wallets             string `yaml:"wallets"`
	WalletTransactions  string `yaml:"wallet_transactions"`
	ProfileTransactions string `yaml:"profile_transactions"`
	Missions            string `yaml:"missions"`
}

// DataConfig holds data output settings
type DataConfig struct {
	OutputDir   string   `yaml:"output_dir"`
	Formats     []string `yaml:"formats"`
	Compression bool     `yaml:"compression"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level     string `yaml:"level"`
	Output    string `yaml:"output"`
	LogToFile bool   `yaml:"log_to_file"`
	LogDir    string `yaml:"log_dir"`
}

// OpenAIConfig holds OpenAI API settings
type OpenAIConfig struct {
	Model          string  `yaml:"model"`
	MaxTokens      int     `yaml:"max_tokens"`
	Temperature    float64 `yaml:"temperature"`
	TimeoutSeconds int     `yaml:"timeout_seconds"`
}

// PromptsConfig holds prompt template settings
type PromptsConfig struct {
	TemplateFile      string `yaml:"template_file"`
	SystemMessageFile string `yaml:"system_message_file"`
	Week              string `yaml:"week"`
}

// BatchConfig holds batch processing settings
type BatchConfig struct {
	Size          int `yaml:"size"`
	MaxConcurrent int `yaml:"max_concurrent"`
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
}

// RetryConfig holds retry settings
type RetryConfig struct {
	MaxAttempts         int  `yaml:"max_attempts"`
	InitialDelaySeconds int  `yaml:"initial_delay_seconds"`
	MaxDelaySeconds     int  `yaml:"max_delay_seconds"`
	ExponentialBackoff  bool `yaml:"exponential_backoff"`
}

// FormattingConfig holds table formatting settings
type FormattingConfig struct {
	EnableTable        bool `yaml:"enable_table"`
	TableWidth         int  `yaml:"table_width"`
	ShowDetailedErrors bool `yaml:"show_detailed_errors"`
}

// MonitoringConfig holds monitoring flags
type MonitoringConfig struct {
	TrackTokenUsage bool `yaml:"track_token_usage"`
	TrackTiming     bool `yaml:"track_timing"`
	ShowProgress    bool `yaml:"show_progress"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// ConnectionString returns PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}
