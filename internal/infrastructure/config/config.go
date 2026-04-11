package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	App      AppConfig      `yaml:"app"`
	Engine   EngineConfig   `yaml:"engine"`
	HTTP     HTTPConfig     `yaml:"http"`
	Security SecurityConfig `yaml:"security"`
}

type AppConfig struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	LogLevel string `yaml:"log_level"`
}

type EngineConfig struct {
	Mode        string
	RateLimit   int    `yaml:"rate_limit"`
	Concurrency int    `yaml:"concurrency"`
	Timeout     int    `yaml:"timeout_sec"`
	Sleep       int    `yaml:"sleep_ms"`
	UserAgent   string `yaml:"user_agent"`
}

type HTTPConfig struct {
	FollowRedirects bool     `yaml:"follow_redirects"`
	VerifyTLS       bool     `yaml:"verify_tls"`
	Proxy           string   `yaml:"proxy,omitempty"`
	ProxyList       []string `yaml:"proxy_list,omitempty"`
}

type SecurityConfig struct {
	SkipCertVerify bool `yaml:"skip_cert_verify"`
}

func DefaultConfig() *GlobalConfig {
	return &GlobalConfig{
		App: AppConfig{
			Name:     "wprecon",
			Version:  "1.0.0",
			LogLevel: "info",
		},
		Engine: EngineConfig{
			Mode:        "normal",
			RateLimit:   50,
			Concurrency: 10,
			Timeout:     10,
			Sleep:       0,
			UserAgent:   "wprecon/1.0",
		},
		HTTP: HTTPConfig{
			FollowRedirects: true,
			VerifyTLS:       true,
		},
		Security: SecurityConfig{
			SkipCertVerify: false,
		},
	}
}

func LoadFromFile(filePath string) (*GlobalConfig, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

func (gc *GlobalConfig) Save(filePath string) error {
	data, err := yaml.Marshal(gc)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (gc *GlobalConfig) ToScanOptions() *ScanOptions {
	return &ScanOptions{
		Mode:            ScanMode(gc.Engine.Mode),
		RateLimit:       gc.Engine.RateLimit,
		Sleep:           time.Duration(gc.Engine.Sleep) * time.Millisecond,
		Concurrency:     gc.Engine.Concurrency,
		Proxy:           gc.HTTP.Proxy,
		Timeout:         time.Duration(gc.Engine.Timeout) * time.Second,
		UserAgent:       gc.Engine.UserAgent,
		FollowRedirects: gc.HTTP.FollowRedirects,
		VerifyTLS:       gc.HTTP.VerifyTLS && !gc.Security.SkipCertVerify,
	}
}

type ScanMode string
type ScanOptions struct {
	Mode            ScanMode
	RateLimit       int
	Sleep           time.Duration
	Concurrency     int
	Proxy           string
	Timeout         time.Duration
	UserAgent       string
	FollowRedirects bool
	VerifyTLS       bool
	TemplatesTags   []string
	SkipTags        []string
}
