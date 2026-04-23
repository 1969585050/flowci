package deployer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/flowci/flowci/pkg/docker"
)

type Deployer struct {
	docker *docker.Client
}

func NewDeployer(dockerClient *docker.Client) *Deployer {
	return &Deployer{docker: dockerClient}
}

type DeployConfig struct {
	ProjectID      string
	ImageTag       string
	ContainerName  string
	Ports          []PortMapping
	EnvVars        map[string]string
	Volumes        []string
	RestartPolicy string
	Replicas       int
}

type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string
}

type DeployResult struct {
	ContainerID string
	Name        string
	Status      string
	Ports       []PortMapping
	StartedAt   time.Time
}

func (d *Deployer) Deploy(ctx context.Context, cfg *DeployConfig) (*DeployResult, error) {
	dockerCli := d.docker.GetCLI()

	resp, err := dockerCli.ContainerCreate(ctx, &container.Config{
		Image:        cfg.ImageTag,
		Env:          envMapToSlice(cfg.EnvVars),
		ExposedPorts: exposedPorts(cfg.Ports),
	}, &container.HostConfig{
		PortBindings:  portBindings(cfg.Ports),
		Mounts:        mounts(cfg.Volumes),
		RestartPolicy: container.RestartPolicy{Name: cfg.RestartPolicy},
	}, nil, nil, cfg.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := dockerCli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return &DeployResult{
		ContainerID: resp.ID,
		Name:        cfg.ContainerName,
		Status:      "running",
		Ports:       cfg.Ports,
		StartedAt:   time.Now(),
	}, nil
}

func (d *Deployer) DeployCompose(ctx context.Context, projectName, composeFile string) error {
	tmpDir, err := os.MkdirTemp("", "flowci-compose-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(composeFile), 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	cmd := exec.Command("docker-compose", "-f", composePath, "-p", projectName, "up", "-d")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker-compose up failed: %w", err)
	}

	return nil
}

func (d *Deployer) Rollback(ctx context.Context, projectName string) error {
	dockerCli := d.docker.GetCLI()

	containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if name == "/"+projectName || name == projectName {
				if err := dockerCli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
					return fmt.Errorf("failed to remove container: %w", err)
				}
			}
		}
	}

	return nil
}

func (d *Deployer) GetStatus(ctx context.Context, projectName string) ([]*DeployResult, error) {
	dockerCli := d.docker.GetCLI()

	containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var results []*DeployResult
	for _, c := range containers {
		for _, name := range c.Names {
			if name == "/"+projectName || name == projectName {
				results = append(results, &DeployResult{
					ContainerID: c.ID,
					Name:        name,
					Status:      c.State,
					StartedAt:   c.Created,
				})
			}
		}
	}

	return results, nil
}

func (d *Deployer) ListContainers(ctx context.Context) ([]*DeployResult, error) {
	dockerCli := d.docker.GetCLI()

	containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var results []*DeployResult
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
		}
		results = append(results, &DeployResult{
			ContainerID: c.ID,
			Name:        name,
			Status:      c.State,
			StartedAt:   c.Created,
		})
	}

	return results, nil
}

func envMapToSlice(env map[string]string) []string {
	var result []string
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

func exposedPorts(ports []PortMapping) map[types.Port]struct{} {
	result := make(map[types.Port]struct{})
	for _, p := range ports {
		result[types.Port(fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol))] = struct{}{}
	}
	return result
}

func portBindings(ports []PortMapping) map[types.Port][]types.PortBinding {
	result := make(map[types.Port][]types.PortBinding)
	for _, p := range ports {
		result[types.Port(fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol))] = []types.PortBinding{
			{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", p.HostPort)},
		}
	}
	return result
}

func mounts(volumes []string) []mount.Mount {
	var result []mount.Mount
	for _, v := range volumes {
		parts := splitVolume(v)
		if len(parts) >= 2 {
			result = append(result, mount.Mount{
				Type:   mount.TypeBind,
				Source: parts[0],
				Target: parts[1],
			})
		}
	}
	return result
}

func splitVolume(v string) []string {
	for i := 0; i < len(v); i++ {
		if v[i] == ':' {
			return []string{v[:i], v[i+1:]}
		}
	}
	return []string{v}
}
