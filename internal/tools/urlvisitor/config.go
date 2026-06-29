package urlvisitor

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultRatePerSecond = 10
	defaultConcurrency   = 10
	defaultTimeout       = 10 * time.Second
	defaultDuration      = 0
	defaultMaxRate       = 10
	defaultStrategy      = "round_robin"
)

// Config 描述一次已授权的 URL 访问测试。RatePerSecond 是整个工具的全局速率，
// 不是每个 worker 各自的速率。
type Config struct {
	URL             string            `yaml:"url"`
	URLs            []string          `yaml:"urls"`
	Strategy        string            `yaml:"strategy"`
	Method          string            `yaml:"method"`
	RatePerSecond   int               `yaml:"rate_per_second"`
	Concurrency     int               `yaml:"concurrency"`
	Duration        time.Duration     `yaml:"duration"`
	Timeout         time.Duration     `yaml:"timeout"`
	FollowRedirects bool              `yaml:"follow_redirects"`
	Headers         map[string]string `yaml:"headers"`
	UserAgents      []string          `yaml:"user_agents"`
	Safety          SafetyConfig      `yaml:"safety"`
}

type SafetyConfig struct {
	RequireAuthorizationConfirm bool     `yaml:"require_authorization_confirm"`
	MaxRatePerSecond            int      `yaml:"max_rate_per_second"`
	AllowedHosts                []string `yaml:"allowed_hosts"`
	AllowPrivateNetworks        bool     `yaml:"allow_private_networks"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("读取配置失败：%w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("解析配置失败：%w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func DefaultConfig() Config {
	return Config{
		Method:          "GET",
		Strategy:        defaultStrategy,
		RatePerSecond:   defaultRatePerSecond,
		Concurrency:     defaultConcurrency,
		Duration:        defaultDuration,
		Timeout:         defaultTimeout,
		FollowRedirects: true,
		Headers:         make(map[string]string),
		UserAgents:      defaultUserAgents(),
		Safety: SafetyConfig{
			RequireAuthorizationConfirm: true,
			MaxRatePerSecond:            defaultMaxRate,
		},
	}
}

func (c Config) Validate() error {
	targets := c.TargetURLs()
	if len(targets) == 0 {
		return errors.New("缺少必填配置：url 或 urls")
	}
	if c.URL != "" && len(c.URLs) > 0 {
		return errors.New("url 和 urls 不能同时配置，请只选择一种")
	}
	if c.Method == "" {
		return errors.New("缺少必填配置：method")
	}
	if c.RatePerSecond <= 0 {
		return errors.New("rate_per_second 必须大于 0")
	}
	if c.Concurrency <= 0 {
		return errors.New("concurrency 必须大于 0")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout 必须大于 0")
	}
	if c.Safety.MaxRatePerSecond <= 0 {
		return errors.New("safety.max_rate_per_second 必须大于 0")
	}
	if c.RatePerSecond > c.Safety.MaxRatePerSecond {
		return fmt.Errorf("rate_per_second=%d 超过安全上限 safety.max_rate_per_second=%d", c.RatePerSecond, c.Safety.MaxRatePerSecond)
	}
	if c.Strategy != "round_robin" && c.Strategy != "random" {
		return fmt.Errorf("不支持的访问策略 %q；当前只支持 round_robin 和 random", c.Strategy)
	}
	return nil
}

func (c Config) TargetURLs() []string {
	if c.URL != "" {
		return []string{c.URL}
	}
	return append([]string(nil), c.URLs...)
}
