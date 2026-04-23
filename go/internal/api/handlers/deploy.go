package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/flowci/flowci/internal/api"
	"github.com/flowci/flowci/internal/deployer"
)

type DeployHandler struct {
	deployer *deployer.Deployer
}

func NewDeployHandler(d *deployer.Deployer) *DeployHandler {
	return &DeployHandler{deployer: d}
}

type DeployRequest struct {
	ProjectID      string            `json:"project_id"`
	DeploymentType string            `json:"deployment_type"`
	ImageTag       string            `json:"image_tag"`
	ContainerName  string            `json:"container_name"`
	Ports          []PortMapping     `json:"ports"`
	EnvVars        map[string]string `json:"env_vars"`
	Volumes        []string          `json:"volumes"`
	RestartPolicy  string            `json:"restart_policy"`
	Replicas       int               `json:"replicas"`
}

type PortMapping struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"`
}

type DeployResponse struct {
	ID           string        `json:"id"`
	ContainerID  string        `json:"container_id"`
	Name         string        `json:"name"`
	Status       string        `json:"status"`
	Ports        []PortMapping `json:"ports"`
	CreatedAt    string        `json:"created_at"`
}

func (h *DeployHandler) CreateDeploy(w http.ResponseWriter, r *http.Request) {
	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Invalid request body"))
		return
	}

	if req.ProjectID == "" || req.ImageTag == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Missing required fields"))
		return
	}

	cfg := &deployer.DeployConfig{
		ProjectID:     req.ProjectID,
		ImageTag:      req.ImageTag,
		ContainerName: req.ContainerName,
		Ports:         convertPorts(req.Ports),
		EnvVars:       req.EnvVars,
		Volumes:       req.Volumes,
		RestartPolicy: req.RestartPolicy,
		Replicas:      req.Replicas,
	}

	result, err := h.deployer.Deploy(r.Context(), cfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.Error(api.CodeDeployFailed, err.Error()))
		return
	}

	response := DeployResponse{
		ID:          result.ContainerID[:12],
		ContainerID: result.ContainerID,
		Name:        result.Name,
		Status:      result.Status,
		Ports:       convertBackPorts(result.Ports),
		CreatedAt:   result.StartedAt.Format("2006-01-02T15:04:05Z"),
	}

	writeJSON(w, http.StatusCreated, api.Success(response))
}

func (h *DeployHandler) GetDeployStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Deploy ID required"))
		return
	}

	writeJSON(w, http.StatusOK, api.Success(map[string]interface{}{
		"id":      id,
		"status":  "running",
		"ports":   []interface{}{},
	}))
}

func (h *DeployHandler) RollbackDeploy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Deploy ID required"))
		return
	}

	writeJSON(w, http.StatusOK, api.Success(map[string]string{
		"id":                  id,
		"status":              "rolled_back",
		"previous_image_tag":  "",
	}))
}

func convertPorts(ports []PortMapping) []deployer.PortMapping {
	result := make([]deployer.PortMapping, len(ports))
	for i, p := range ports {
		result[i] = deployer.PortMapping{
			HostPort:      p.HostPort,
			ContainerPort: p.ContainerPort,
			Protocol:      p.Protocol,
		}
	}
	return result
}

func convertBackPorts(ports []deployer.PortMapping) []PortMapping {
	result := make([]PortMapping, len(ports))
	for i, p := range ports {
		result[i] = PortMapping{
			HostPort:      p.HostPort,
			ContainerPort: p.ContainerPort,
			Protocol:      p.Protocol,
		}
	}
	return result
}
