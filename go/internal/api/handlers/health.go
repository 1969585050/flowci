package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flowci/flowci/internal/api"
	"github.com/flowci/flowci/pkg/docker"
)

type HealthHandler struct {
	dockerClient *docker.Client
	version      string
	startTime    time.Time
}

func NewHealthHandler(dc *docker.Client, version string) *HealthHandler {
	return &HealthHandler{
		dockerClient: dc,
		version:      version,
		startTime:    time.Now(),
	}
}

type HealthResponse struct {
	Status          string `json:"status"`
	Version         string `json:"version"`
	UptimeSeconds   int64  `json:"uptime_seconds"`
	DockerConnected bool   `json:"docker_connected"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	connected := false
	if h.dockerClient != nil {
		cli := h.dockerClient.GetCLI()
		_, err := cli.Ping(r.Context())
		connected = err == nil
	}

	response := HealthResponse{
		Status:          "healthy",
		Version:         h.version,
		UptimeSeconds:   int64(time.Since(h.startTime).Seconds()),
		DockerConnected: connected,
	}

	status := http.StatusOK
	if !connected {
		response.Status = "degraded"
	}

	writeJSON(w, status, api.Success(response))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
