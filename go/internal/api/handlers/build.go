package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/flowci/flowci/internal/api"
	"github.com/flowci/flowci/internal/builder"
	"github.com/flowci/flowci/pkg/docker"
)

type BuildHandler struct {
	builder     *builder.Builder
	dockerClient *docker.Client
}

func NewBuildHandler(b *builder.Builder, dc *docker.Client) *BuildHandler {
	return &BuildHandler{
		builder:     b,
		dockerClient: dc,
	}
}

type CreateBuildRequest struct {
	ProjectID     string            `json:"project_id" validate:"required"`
	Language      string            `json:"language" validate:"required"`
	ContextPath  string            `json:"context_path" validate:"required"`
	DockerfilePath string          `json:"dockerfile_path,omitempty"`
	ImageTags    []string          `json:"image_tags" validate:"required,min=1"`
	BuildArgs    map[string]string `json:"build_args,omitempty"`
	NoCache      bool              `json:"no_cache,omitempty"`
	PullBaseImage bool            `json:"pull_base_image,omitempty"`
}

type BuildResponse struct {
	ID          string   `json:"id"`
	ImageID     string   `json:"image_id"`
	Tags        []string `json:"tags"`
	Size        int64    `json:"size"`
	DurationMs  int64    `json:"duration_ms"`
	Status      string   `json:"status"`
	Logs        []string `json:"logs,omitempty"`
	StartedAt   string   `json:"started_at"`
	FinishedAt  string   `json:"finished_at"`
}

type BuildListResponse struct {
	Builds []BuildResponse `json:"builds"`
	Total  int            `json:"total"`
}

type BuildLogsResponse struct {
	BuildID string   `json:"build_id"`
	Logs    []string `json:"logs"`
}

func (h *BuildHandler) CreateBuild(w http.ResponseWriter, r *http.Request) {
	var req CreateBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Invalid request body: "+err.Error()))
		return
	}

	if err := validateCreateBuildRequest(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, err.Error()))
		return
	}

	if err := builder.ValidateLanguage(req.Language); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, err.Error()))
		return
	}

	cfg := &builder.BuildConfig{
		ProjectID:      req.ProjectID,
		Language:       builder.Language(req.Language),
		ContextPath:   req.ContextPath,
		DockerfilePath: req.DockerfilePath,
		ImageTags:     req.ImageTags,
		BuildArgs:     req.BuildArgs,
		NoCache:       req.NoCache,
		PullBaseImage:  req.PullBaseImage,
	}

	result, err := h.builder.Build(r.Context(), cfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorWithData(
			api.CodeBuildFailed,
			"Build failed: "+err.Error(),
			map[string]interface{}{
				"project_id": req.ProjectID,
				"language":   req.Language,
			},
		))
		return
	}

	response := BuildResponse{
		ID:         result.ID,
		ImageID:    result.ImageID,
		Tags:       result.ImageTags,
		Size:       result.Size,
		DurationMs: result.Duration.Milliseconds(),
		Status:     string(result.Status),
		Logs:       result.Logs,
		StartedAt:  result.StartedAt.Format(time.RFC3339),
		FinishedAt: result.FinishedAt.Format(time.RFC3339),
	}

	writeJSON(w, http.StatusCreated, api.Success(response))
}

func (h *BuildHandler) GetBuild(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Build ID is required"))
		return
	}

	writeJSON(w, http.StatusOK, api.Success(BuildResponse{
		ID:      id,
		Status:  "not_implemented",
	}))
}

func (h *BuildHandler) GetBuildLogs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, api.Error(api.CodeInvalidParam, "Build ID is required"))
		return
	}

	writeJSON(w, http.StatusOK, api.Success(BuildLogsResponse{
		BuildID: id,
		Logs:    []string{},
	}))
}

func (h *BuildHandler) ListBuilds(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, api.Success(BuildListResponse{
		Builds: []BuildResponse{},
		Total:  0,
	}))
}

func validateCreateBuildRequest(req *CreateBuildRequest) error {
	if req.ProjectID == "" {
		return &ValidationError{Field: "project_id", Message: "project_id is required"}
	}
	if req.Language == "" {
		return &ValidationError{Field: "language", Message: "language is required"}
	}
	if req.ContextPath == "" {
		return &ValidationError{Field: "context_path", Message: "context_path is required"}
	}
	if len(req.ImageTags) == 0 {
		return &ValidationError{Field: "image_tags", Message: "at least one image_tag is required"}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
