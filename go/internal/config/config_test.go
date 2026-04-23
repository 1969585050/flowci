package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	assert.Equal(t, "0.1.0", cfg.Version)
	assert.NotNil(t, cfg.Projects)
	assert.NotNil(t, cfg.Registries)
	assert.NotNil(t, cfg.Credentials)
	assert.Empty(t, cfg.Projects)
}

func TestNewManager(t *testing.T) {
	cfg := Default()
	mgr := NewManager(cfg)

	assert.NotNil(t, mgr)
	assert.Equal(t, cfg, mgr.cfg)
}

func TestManager_GetProjects(t *testing.T) {
	cfg := Default()
	cfg.Projects = []*Project{
		{ID: "proj-1", Name: "Project 1"},
		{ID: "proj-2", Name: "Project 2"},
	}

	mgr := NewManager(cfg)
	projects := mgr.GetProjects()

	assert.Len(t, projects, 2)
	assert.Equal(t, "proj-1", projects[0].ID)
}

func TestManager_GetRegistries(t *testing.T) {
	cfg := Default()
	cfg.Registries = []*Registry{
		{ID: "reg-1", Name: "Registry 1"},
	}

	mgr := NewManager(cfg)
	registries := mgr.GetRegistries()

	assert.Len(t, registries, 1)
	assert.Equal(t, "reg-1", registries[0].ID)
}

func TestManager_GetCredentials(t *testing.T) {
	cfg := Default()
	cfg.Credentials = []*Credential{
		{ID: "cred-1", Name: "Credential 1"},
	}

	mgr := NewManager(cfg)
	creds := mgr.GetCredentials()

	assert.Len(t, creds, 1)
	assert.Equal(t, "cred-1", creds[0].ID)
}

func TestGetConfigPath(t *testing.T) {
	path, err := getConfigPath()

	assert.NoError(t, err)
	assert.Contains(t, path, ".config")
	assert.Contains(t, path, "flowci")
	assert.Contains(t, path, "config.yaml")
}

func TestProject_Structure(t *testing.T) {
	project := Project{
		ID:       "test-id",
		Name:     "Test Project",
		Path:     "/path/to/project",
		Language: "go",
		BuildConfig: map[string]interface{}{
			"dockerfile": "Dockerfile",
		},
		DeployConfig: map[string]interface{}{
			"replicas": 3,
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-02T00:00:00Z",
	}

	assert.Equal(t, "test-id", project.ID)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, "/path/to/project", project.Path)
	assert.Equal(t, "go", project.Language)
	assert.Equal(t, "Dockerfile", project.BuildConfig["dockerfile"])
	assert.Equal(t, 3, project.DeployConfig["replicas"])
}

func TestRegistry_Structure(t *testing.T) {
	registry := Registry{
		ID:           "reg-1",
		Name:         "Docker Hub",
		RegistryType: "dockerhub",
		Address:      "registry.hub.docker.com",
		Namespace:    "namespace",
		URL:         "https://registry.hub.docker.com",
	}

	assert.Equal(t, "reg-1", registry.ID)
	assert.Equal(t, "Docker Hub", registry.Name)
	assert.Equal(t, "dockerhub", registry.RegistryType)
}

func TestCredential_Structure(t *testing.T) {
	cred := Credential{
		ID:            "cred-1",
		Name:          "Docker Auth",
		CredentialType: "docker",
		Username:      "user",
		Password:      "pass",
		SSHKey:       "",
	}

	assert.Equal(t, "cred-1", cred.ID)
	assert.Equal(t, "Docker Auth", cred.Name)
	assert.Equal(t, "docker", cred.CredentialType)
	assert.Equal(t, "user", cred.Username)
	assert.Equal(t, "pass", cred.Password)
}

func TestDockerConfig_Structure(t *testing.T) {
	dockerCfg := DockerConfig{
		SocketPath: "/var/run/docker.sock",
		Host:       "localhost",
		Port:       2375,
		UseSSH:     true,
	}

	assert.Equal(t, "/var/run/docker.sock", dockerCfg.SocketPath)
	assert.Equal(t, "localhost", dockerCfg.Host)
	assert.Equal(t, 2375, dockerCfg.Port)
	assert.True(t, dockerCfg.UseSSH)
}

func TestConfig_Structure(t *testing.T) {
	cfg := &Config{
		Version: "1.0.0",
		Docker: DockerConfig{
			SocketPath: "/var/run/docker.sock",
		},
		Projects:    []*Project{},
		Registries:  []*Registry{},
		Credentials: []*Credential{},
	}

	assert.Equal(t, "1.0.0", cfg.Version)
	assert.Equal(t, "/var/run/docker.sock", cfg.Docker.SocketPath)
	assert.Empty(t, cfg.Projects)
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		Version: "test-version",
		Docker: DockerConfig{
			Host: "test-host",
			Port: 2375,
		},
		Projects: []*Project{
			{
				ID:       "proj-1",
				Name:     "Test Project",
				Language: "go",
			},
		},
	}

	data, err := yaml.Marshal(cfg)
	assert.NoError(t, err)

	err = os.WriteFile(configPath, data, 0644)
	assert.NoError(t, err)

	loadedData, err := os.ReadFile(configPath)
	assert.NoError(t, err)

	var loadedCfg Config
	err = yaml.Unmarshal(loadedData, &loadedCfg)
	assert.NoError(t, err)

	assert.Equal(t, "test-version", loadedCfg.Version)
	assert.Equal(t, "test-host", loadedCfg.Docker.Host)
	assert.Len(t, loadedCfg.Projects, 1)
	assert.Equal(t, "proj-1", loadedCfg.Projects[0].ID)
}
