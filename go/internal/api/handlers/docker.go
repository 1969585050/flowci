package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flowci/flowci/internal/api"
	"github.com/flowci/flowci/pkg/docker"
)

type DockerHandler struct {
	dockerClient *docker.Client
}

func NewDockerHandler(dc *docker.Client) *DockerHandler {
	return &DockerHandler{dockerClient: dc}
}

type DockerCheckResponse struct {
	Connected   bool   `json:"connected"`
	Version     string `json:"version"`
	APIVersion  string `json:"api_version"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
}

type ImageInfo struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Size    int64    `json:"size"`
	Created string   `json:"created"`
}

type ContainerInfo struct {
	ID      string        `json:"id"`
	Names   []string      `json:"names"`
	Image   string        `json:"image"`
	State   string        `json:"state"`
	Status  string        `json:"status"`
	Ports   []PortMapping `json:"ports"`
	Created string        `json:"created"`
}

func (h *DockerHandler) CheckConnection(w http.ResponseWriter, r *http.Request) {
	cli := h.dockerClient.GetCLI()
	info, err := cli.Info(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, api.Success(DockerCheckResponse{
			Connected: false,
			Version:   "disconnected",
		}))
		return
	}

	writeJSON(w, http.StatusOK, api.Success(DockerCheckResponse{
		Connected:  true,
		Version:    info.ServerVersion,
		APIVersion: cli.ClientVersion(),
		OS:         info.OperatingSystem,
		Arch:       info.Architecture,
	}))
}

func (h *DockerHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	images, err := h.dockerClient.ListImages(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.Error(api.CodeDockerConnFailed, err.Error()))
		return
	}

	imageList := make([]ImageInfo, len(images))
	for i, img := range images {
		tags := img.RepoTags
		if tags == nil {
			tags = []string{"<none>:<none>"}
		}
		imageList[i] = ImageInfo{
			ID:      img.ID,
			Tags:    tags,
			Size:    img.Size,
			Created: img.Created.String(),
		}
	}

	writeJSON(w, http.StatusOK, api.Success(map[string]interface{}{
		"images": imageList,
		"total":  len(imageList),
	}))
}

func (h *DockerHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := h.dockerClient.ListContainers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.Error(api.CodeDockerConnFailed, err.Error()))
		return
	}

	containerList := make([]ContainerInfo, len(containers))
	for i, c := range containers {
		ports := make([]PortMapping, len(c.Ports))
		for j, p := range c.Ports {
			ports[j] = PortMapping{
				HostPort:      int(p.PublicPort),
				ContainerPort: int(p.PrivatePort),
				Protocol:      p.Type,
			}
		}
		containerList[i] = ContainerInfo{
			ID:      c.ID,
			Names:   c.Names,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Ports:   ports,
			Created: c.Created.String(),
		}
	}

	writeJSON(w, http.StatusOK, api.Success(map[string]interface{}{
		"containers": containerList,
		"total":      len(containerList),
	}))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
