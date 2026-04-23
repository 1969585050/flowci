package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/flowci/flowci/internal/api"
	"github.com/flowci/flowci/internal/config"
)

type ProjectHandler struct {
	configMgr *config.Manager
}

func NewProjectHandler(cm *config.Manager) *ProjectHandler {
	return &ProjectHandler{configMgr: cm}
}

type CreateProjectRequest struct {
	Name         string                 `json:"name"`
	Path         string                 `json:"path"`
	Language     string                 `json:"language"`
	BuildConfig  map[string]interface{} `json:"build_config"`
	DeployConfig map[string]interface{} `json:"deploy_config"`
}

type UpdateProjectRequest struct {
	Name         string                 `json:"name"`
	Language     string                 `json:"language"`
	BuildConfig  map[string]interface{} `json:"build_config"`
	DeployConfig map[string]interface{} `json:"deploy_config"`
}

type ProjectResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Path          string                 `json:"path"`
	Language      string                 `json:"language"`
	BuildConfig   map[string]interface{} `json:"build_config,omitempty"`
	DeployConfig  map[string]interface{} `json:"deploy_config,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects := h.configMgr.GetProjects()

	projectList := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		projectList[i] = ProjectResponse{
			ID:        p.ID,
			Name:      p.Name,
			Path:      p.Path,
			Language:  p.Language,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
	}

	writeJSON(w, http.StatusOK, api.Success(map[string]interface{}{
		"projects": projectList,
		"total":    len(projectList),
	}))
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Invalid request body"))
		return
	}

	if req.Name == "" || req.Path == "" || req.Language == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Missing required fields"))
		return
	}

	cfg := h.configMgr.GetConfig()
	newProject := &config.Project{
		ID:           generateID(),
		Name:         req.Name,
		Path:         req.Path,
		Language:     req.Language,
		BuildConfig:  req.BuildConfig,
		DeployConfig: req.DeployConfig,
	}

	cfg.Projects = append(cfg.Projects, newProject)
	if err := config.Save(cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, api.Error(api.CodeInternalError, err.Error()))
		return
	}

	response := ProjectResponse{
		ID:        newProject.ID,
		Name:      newProject.Name,
		Path:      newProject.Path,
		Language:  newProject.Language,
		CreatedAt: newProject.CreatedAt,
		UpdatedAt: newProject.UpdatedAt,
	}

	writeJSON(w, http.StatusCreated, api.Success(response))
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Project ID required"))
		return
	}

	projects := h.configMgr.GetProjects()
	for _, p := range projects {
		if p.ID == id {
			writeJSON(w, http.StatusOK, api.Success(ProjectResponse{
				ID:           p.ID,
				Name:         p.Name,
				Path:         p.Path,
				Language:     p.Language,
				BuildConfig:  p.BuildConfig,
				DeployConfig: p.DeployConfig,
				CreatedAt:    p.CreatedAt,
				UpdatedAt:    p.UpdatedAt,
			}))
			return
		}
	}

	writeJSON(w, http.StatusNotFound, api.Error(api.CodeNotFound, "Project not found"))
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Project ID required"))
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Invalid request body"))
		return
	}

	cfg := h.configMgr.GetConfig()
	var target *config.Project
	for _, p := range cfg.Projects {
		if p.ID == id {
			target = p
			break
		}
	}

	if target == nil {
		writeJSON(w, http.StatusNotFound, api.Error(api.CodeNotFound, "Project not found"))
		return
	}

	if req.Name != "" {
		target.Name = req.Name
	}
	if req.Language != "" {
		target.Language = req.Language
	}
	if req.BuildConfig != nil {
		target.BuildConfig = req.BuildConfig
	}
	if req.DeployConfig != nil {
		target.DeployConfig = req.DeployConfig
	}

	if err := config.Save(cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, api.Error(api.CodeInternalError, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, api.Success(map[string]interface{}{
		"id":         target.ID,
		"name":       target.Name,
		"updated_at": target.UpdatedAt,
	}))
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Project ID required"))
		return
	}

	cfg := h.configMgr.GetConfig()
	found := false
	var newProjects []*config.Project
	for _, p := range cfg.Projects {
		if p.ID == id {
			found = true
		} else {
			newProjects = append(newProjects, p)
		}
	}

	if !found {
		writeJSON(w, http.StatusNotFound, api.Error(api.CodeNotFound, "Project not found"))
		return
	}

	cfg.Projects = newProjects
	if err := config.Save(cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, api.Error(api.CodeInternalError, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, api.Success(nil))
}

func generateID() string {
	return "proj-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
