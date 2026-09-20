package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Profile struct {
	Datacenter string        `yaml:"datacenter"`
	AuthMethod string        `yaml:"auth_method"`
	Token      string        `yaml:"token"`
	BaseURL    string        `yaml:"base_url"`
	Timeout    time.Duration `yaml:"timeout"`
	ReadOnly   bool          `yaml:"read_only"`
}

type Config struct {
	DefaultProfile string              `yaml:"default_profile"`
	Profiles       map[string]*Profile `yaml:"profiles"`

	ProfileName string   `yaml:"-"`
	Active      *Profile `yaml:"-"`
}

func defaults() *Config {
	return &Config{
		DefaultProfile: "default",
		Profiles: map[string]*Profile{
			"default": {
				AuthMethod: "token",
				Timeout:    30 * time.Second,
			},
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := defaults()

	if path == "" {
		path = DefaultConfigPath()
	}
	if err := applyFileFromPath(cfg, path); err != nil {
		return cfg, err
	}
	cfg.resolve()
	return cfg, nil
}

func applyFileFromPath(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config %s: %w", path, err)
	}
	raw := map[string]any{}
	if err := yamlUnmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	applyFile(cfg, raw)
	return nil
}

func applyFile(cfg *Config, raw map[string]any) {
	if v, ok := raw["default_profile"].(string); ok && v != "" {
		cfg.DefaultProfile = v
	}
	profiles, ok := raw["profiles"].(map[string]any)
	if !ok {
		return
	}
	for name, p := range profiles {
		fields, ok := p.(map[string]any)
		if !ok {
			continue
		}
		prof := cfg.ensureProfile(name)
		if v, ok := fields["datacenter"].(string); ok {
			prof.Datacenter = v
		}
		if v, ok := fields["auth_method"].(string); ok {
			prof.AuthMethod = v
		}
		if v, ok := fields["token"].(string); ok {
			prof.Token = v
		}
		if v, ok := fields["base_url"].(string); ok {
			prof.BaseURL = v
		}
		if v, ok := fields["timeout"]; ok {
			if d, dok := coerceDuration(v); dok {
				prof.Timeout = d
			}
		}
		if v, ok := fields["read_only"].(bool); ok {
			prof.ReadOnly = v
		}
	}
}

func (c *Config) ensureProfile(name string) *Profile {
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	p, ok := c.Profiles[name]
	if !ok {
		p = &Profile{AuthMethod: "token", Timeout: 30 * time.Second}
		c.Profiles[name] = p
	}
	return p
}

func (c *Config) resolve() {
	name := c.DefaultProfile
	if v := os.Getenv("QUALTRICS_PROFILE"); v != "" {
		name = v
	}
	if name == "" {
		name = "default"
	}
	c.ProfileName = name

	prof := c.ensureProfile(name)
	if v := os.Getenv("QUALTRICS_DATACENTER"); v != "" {
		prof.Datacenter = v
	}
	if v := os.Getenv("QUALTRICS_BASE_URL"); v != "" {
		prof.BaseURL = v
	}
	if v := os.Getenv("QUALTRICS_TOKEN"); v != "" {
		prof.Token = v
	}
	if v := os.Getenv("QUALTRICS_AUTH_METHOD"); v != "" {
		prof.AuthMethod = v
	}
	if v := os.Getenv("QUALTRICS_TIMEOUT"); v != "" {
		if d, ok := coerceDuration(v); ok {
			prof.Timeout = d
		}
	}
	if v := os.Getenv("QUALTRICS_READ_ONLY"); v != "" {
		prof.ReadOnly = ParseBool(v)
	}
	c.Active = prof
}

func (c *Config) Token() string {
	if c.Active == nil {
		return ""
	}
	return ResolveSecret(c.Active.Token)
}

func (c *Config) BaseURL() (string, error) {
	if c.Active == nil {
		return "", fmt.Errorf("no active profile")
	}
	if c.Active.BaseURL != "" {
		return strings.TrimRight(c.Active.BaseURL, "/"), nil
	}
	if c.Active.Datacenter != "" {
		return "https://" + c.Active.Datacenter + ".qualtrics.com/API/v3", nil
	}
	return "", fmt.Errorf("profile %q has no datacenter or base_url; set one with `qualtrics auth set-datacenter <id>`", c.ProfileName)
}

func coerceDuration(v any) (time.Duration, bool) {
	switch t := v.(type) {
	case string:
		if d, err := time.ParseDuration(t); err == nil {
			return d, true
		}
		if n, err := parseInt(t); err == nil {
			return time.Duration(n), true
		}
	case int:
		return time.Duration(t), true
	case int64:
		return time.Duration(t), true
	case float64:
		return time.Duration(int64(t)), true
	}
	return 0, false
}

func ParseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}
