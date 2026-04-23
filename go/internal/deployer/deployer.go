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
	"github.com/flowci/flowci/pkg/docker"
)

type Deployer struct {
	docker *docker.Client
}

func NewDeployer(dockerClient *docker.Client) *Deployer {
	return &Deployer{docker: dockerClient}
}

type DeployConfig struct {
	ProjectID     string
	ImageTag      string
	ContainerName string
	Ports         []PortMapping
	EnvVars       map[string]string
	Volumes       []string
	RestartPolicy string
	Replicas      int
}

type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string
}

type DeployResult struct {
	ID           string
	ContainerID  string
	Name         string
	Image        string
	Status       string
	Ports        []PortMapping
	StartedAt    time.Time
}

type DeployStatus string

const (
	DeployStatusPending   DeployStatus = "pending"
	DeployStatusRunning   DeployStatus = "running"
	DeployStatusStopped  DeployStatus = "stopped"
	DeployStatusFailed   DeployStatus = "failed"
	DeployStatusRolling  DeployStatus = "rolling_update"
)

func (d *Deployer) Deploy(ctx context.Context, cfg *DeployConfig) (*DeployResult, error) {
	cli := d.docker.GetCLI()

	containerName := cfg.ContainerName
	if containerName == "" {
		containerName = fmt.Sprintf("flowci-%s", cfg.ProjectID)
	}

	envSlice := envMapToSlice(cfg.EnvVars)
	exposedPorts := exposedPorts(cfg.Ports)
	portBindings := portBindings(cfg.Ports)
	mounts := mounts(cfg.Volumes)

	restartPolicy := container.RestartPolicy{
		Name: cfg.RestartPolicy,
	}
	if restartPolicy.Name == "" {
		restartPolicy.Name = "unless-stopped"
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        cfg.ImageTag,
		Env:          envSlice,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings:  portBindings,
		Mounts:       mounts,
		RestartPolicy: restartPolicy,
	}, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return &DeployResult{
		ID:          generateDeployID(),
		ContainerID: resp.ID,
		Name:        containerName,
		Image:       cfg.ImageTag,
		Status:      string(DeployStatusRunning),
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
	cli := d.docker.GetCLI()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			cleanName := trimSlash(name)
			if cleanName == projectName || cleanName == "/"+projectName {
				if err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
					return fmt.Errorf("failed to remove container: %w", err)
				}
			}
		}
	}

	return nil
}

func (d *Deployer) GetStatus(ctx context.Context, projectName string) ([]*DeployResult, error) {
	cli := d.docker.GetCLI()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var results []*DeployResult
	for _, c := range containers {
		for _, name := range c.Names {
			cleanName := trimSlash(name)
			if cleanName == projectName || cleanName == "/"+projectName {
				ports := make([]PortMapping, len(c.Ports))
				for i, p := range c.Ports {
					ports[i] = PortMapping{
						HostPort:      int(p.PublicPort),
						ContainerPort: int(p.PrivatePort),
						Protocol:      p.Type,
					}
				}

				results = append(results, &DeployResult{
					ID:          c.ID[:12],
					ContainerID: c.ID,
					Name:        cleanName,
					Image:       c.Image,
					Status:      c.State,
					Ports:       ports,
					StartedAt:   c.Created,
				})
			}
		}
	}

	return results, nil
}

func (d *Deployer) ListContainers(ctx context.Context) ([]*DeployResult, error) {
	cli := d.docker.GetCLI()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var results []*DeployResult
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = trimSlash(c.Names[0])
		}

		ports := make([]PortMapping, len(c.Ports))
		for i, p := range c.Ports {
			ports[i] = PortMapping{
				HostPort:      int(p.PublicPort),
				ContainerPort: int(p.PrivatePort),
				Protocol:      p.Type,
			}
		}

		results = append(results, &DeployResult{
			ID:          c.ID[:12],
			ContainerID: c.ID,
			Name:        name,
			Image:       c.Image,
			Status:      c.State,
			Ports:       ports,
			StartedAt:   c.Created,
		})
	}

	return results, nil
}

func (d *Deployer) StopContainer(ctx context.Context, containerID string) error {
	cli := d.docker.GetCLI()
	timeout := 10
	return cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

func (d *Deployer) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	cli := d.docker.GetCLI()
	return cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: force})
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

func trimSlash(name string) string {
	if len(name) > 0 && name[0] == '/' {
		return name[1:]
	}
	return name
}

func generateDeployID() string {
	return fmt.Sprintf("deploy-%d", time.Now().UnixNano())
}
