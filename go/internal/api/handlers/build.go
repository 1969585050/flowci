package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/flowci/flowci/internal/api"
	"github.com/flowci/flowci/internal/builder"
	"github.com/flowci/flowci/pkg/docker"
)

type BuildHandler struct {
	builder  *builder.Builder
	dockerClient *docker.Client
}

func NewBuildHandler(b *builder.Builder, dc *docker.Client) *BuildHandler {
	return &BuildHandler{
		builder:    b,
		dockerClient: dc,
	}
}

type CreateBuildRequest struct {
	ProjectID     string            `json:"project_id"`
	Language      string            `json:"language"`
	ContextPath   string            `json:"context_path"`
	ImageTags     []string          `json:"image_tags"`
	BuildArgs     map[string]string `json:"build_args"`
	NoCache       bool              `json:"no_cache"`
	PullBaseImage bool              `json:"pull_base_image"`
}

type BuildResponse struct {
	ID         string   `json:"id"`
	ImageID    string   `json:"image_id"`
	Tags       []string `json:"tags"`
	Size       int64    `json:"size"`
	DurationMs int64    `json:"duration_ms"`
	Status     string   `json:"status"`
	Logs       []string `json:"logs,omitempty"`
	CreatedAt  string   `json:"created_at"`
}

func (h *BuildHandler) CreateBuild(w http.ResponseWriter, r *http.Request) {
	var req CreateBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Invalid request body"))
		return
	}

	if req.ProjectID == "" || req.Language == "" || req.ContextPath == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Missing required fields"))
		return
	}

	cfg := &builder.BuildConfig{
		ProjectID:      req.ProjectID,
		Language:       builder.Language(req.Language),
		ContextPath:    req.ContextPath,
		ImageTags:      req.ImageTags,
		BuildArgs:      req.BuildArgs,
		NoCache:        req.NoCache,
		PullBaseImage:  req.PullBaseImage,
	}

	result, err := h.builder.Build(r.Context(), cfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.Error(api.CodeBuildFailed, err.Error()))
		return
	}

	response := BuildResponse{
		ID:         result.ImageID[:12],
		ImageID:    result.ImageID,
		Tags:       result.ImageTags,
		Size:       result.Size,
		DurationMs: result.Duration.Milliseconds(),
		Status:     "success",
		CreatedAt:  result.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, api.Success(response))
}

func (h *BuildHandler) GetBuild(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Build ID required"))
		return
	}

	writeJSON(w, http.StatusOK, api.Success(map[string]string{
		"id":     id,
		"status": "success",
	}))
}

func (h *BuildHandler) GetBuildLogs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Build ID required"))
		return
	}

	writeJSON(w, http.StatusOK, api.Success(map[string]interface{}{
		"build_id": id,
		"logs":     []interface{}{},
	}))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
