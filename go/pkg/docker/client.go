package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	if _, err := cli.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("docker ping failed: %w", err)
	}

	return &Client{cli: cli}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

func (c *Client) GetCLI() *client.Client {
	return c.cli
}

func (c *Client) ListImages(ctx context.Context) ([]types.ImageSummary, error) {
	return c.cli.ImageList(ctx, types.ImageListOptions{})
}

func (c *Client) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	return c.cli.ContainerList(ctx, types.ContainerListOptions{All: all})
}

func (c *Client) GetContainer(ctx context.Context, id string) (types.Container, error) {
	container, err := c.cli.ContainerInspect(ctx, id)
	if err != nil {
		return types.Container{}, err
	}
	return types.Container{
		ID:      container.ID,
		Names:   []string{container.Name},
		Image:   container.Config.Image,
		State:   container.State.String(),
		Status:  container.State.Status,
		Created: container.Created,
		Ports:   convertPorts(container.NetworkSettings.Ports),
	}, nil
}

func (c *Client) CreateContainer(ctx context.Context, cfg *container.Config, hostCfg *container.HostConfig, name string) (container.CreateResponse, error) {
	return c.cli.ContainerCreate(ctx, cfg, hostCfg, nil, nil, name)
}

func (c *Client) StartContainer(ctx context.Context, id string) error {
	return c.cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

func (c *Client) StopContainer(ctx context.Context, id string, timeout int) error {
	return c.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
}

func (c *Client) RemoveContainer(ctx context.Context, id string, force bool) error {
	return c.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: force})
}

func (c *Client) GetContainerLogs(ctx context.Context, id string, tail string) (string, error) {
	reader, err := c.cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	buf := make([]byte, 1024*1024)
	n, _ := reader.Read(buf)
	return string(buf[:n]), nil
}

func (c *Client) BuildImage(ctx context.Context, buildCtx interface{}, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	return c.cli.ImageBuild(ctx, buildCtx.(interface{ Read([]byte) (int, error) }), options)
}

func (c *Client) PullImage(ctx context.Context, refStr string) error {
	_, err := c.cli.ImagePull(ctx, refStr, types.ImagePullOptions{})
	return err
}

func (c *Client) PushImage(ctx context.Context, ref string, options types.ImagePushOptions) error {
	_, err := c.cli.ImagePush(ctx, ref, options)
	return err
}

func (c *Client) RemoveImage(ctx context.Context, imageID string, options types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error) {
	return c.cli.ImageRemove(ctx, imageID, options)
}

func (c *Client) Info(ctx context.Context) (types.Info, error) {
	return c.cli.Info(ctx)
}

func (c *Client) Version(ctx context.Context) (types.Version, error) {
	return c.cli.ClientVersion(), nil
}

func (c *Client) PruneImages(ctx context.Context) (types.ImagePruneReport, error) {
	return c.cli.ImagesPrune(ctx, filters.NewArgs())
}

func (c *Client) PruneContainers(ctx context.Context) (types.ContainerPruneReport, error) {
	return c.cli.ContainersPrune(ctx, filters.NewArgs())
}

func convertPorts(ports map[types.Port][]types.PortBinding) []types.Port {
	var result []types.Port
	for p, bindings := range ports {
		for _, b := range bindings {
			result = append(result, types.Port{
				PrivatePort: p.Int(),
				PublicPort:  uint16(b.HostPort),
				Type:       p.Proto(),
			})
		}
	}
	return result
}

type CreateContainerRequest struct {
	Name        string
	Image       string
	Env         []string
	ExposedPorts map[types.Port]struct{}
	PortBindings map[types.Port][]types.PortBinding
	Mounts      []container.Mount
	RestartPolicy container.RestartPolicy
}

func (c *Client) CreateAndStartContainer(ctx context.Context, req *CreateContainerRequest) (*container.CreateResponse, error) {
	resp, err := c.cli.ContainerCreate(ctx, &container.Config{
		Image:        req.Image,
		Env:          req.Env,
		ExposedPorts: req.ExposedPorts,
	}, &container.HostConfig{
		PortBindings:  req.PortBindings,
		Mounts:       req.Mounts,
		RestartPolicy: req.RestartPolicy,
	}, nil, nil, req.Name)
	if err != nil {
		return nil, err
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) NetworkCreate(ctx context.Context, name string, driver string) (network.CreateResponse, error) {
	return c.cli.NetworkCreate(ctx, name, types.NetworkCreate{
		Driver: driver,
	})
}

func (c *Client) NetworkConnect(ctx context.Context, networkID, containerID string) error {
	return c.cli.NetworkConnect(ctx, networkID, containerID, &network.EndpointSettings{})
}

func (c *Client) NetworkDisconnect(ctx context.Context, networkID, containerID string) error {
	return c.cli.NetworkDisconnect(ctx, networkID, containerID, true)
}
