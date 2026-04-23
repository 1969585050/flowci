package config

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	cfg *Config
}

func NewManager(cfg *Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) GetProjects() []*Project {
	return m.cfg.Projects
}

func (m *Manager) GetRegistries() []*Registry {
	return m.cfg.Registries
}

func (m *Manager) GetCredentials() []*Credential {
	return m.cfg.Credentials
}

type Config struct {
	Version    string      `yaml:"version"`
	Docker     DockerConfig `yaml:"docker"`
	Projects   []*Project   `yaml:"projects"`
	Registries []*Registry  `yaml:"registries"`
	Credentials []*Credential `yaml:"credentials"`
}

type DockerConfig struct {
	SocketPath string `yaml:"socket_path"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	UseSSH     bool   `yaml:"use_ssh"`
}

type Project struct {
	ID            string                 `yaml:"id"`
	Name          string                 `yaml:"name"`
	Path          string                 `yaml:"path"`
	Language      string                 `yaml:"language"`
	BuildConfig   map[string]interface{} `yaml:"build_config"`
	DeployConfig  map[string]interface{} `yaml:"deploy_config"`
	CreatedAt     string                 `yaml:"created_at"`
	UpdatedAt     string                 `yaml:"updated_at"`
}

type Registry struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	RegistryType string `yaml:"registry_type"`
	Address      string `yaml:"address"`
	Namespace    string `yaml:"namespace"`
	URL          string `yaml:"url"`
}

type Credential struct {
	ID            string `yaml:"id"`
	Name          string `yaml:"name"`
	CredentialType string `yaml:"credential_type"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	SSHKey       string `yaml:"ssh_key"`
}

func Default() *Config {
	return &Config{
		Version:    "0.1.0",
		Docker:     DockerConfig{},
		Projects:   []*Project{},
		Registries: []*Registry{},
		Credentials: []*Credential{},
	}
}

func Load() (*Config, error) {
	cfgPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfg := Default()
		if err := Save(cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	cfgPath, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(cfgPath, data, 0644)
}

func getConfigPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "flowci", "config.yaml"), nil
}
