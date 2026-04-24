package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"flowci/store"
	"gopkg.in/yaml.v3"
)

//go:embed all:dist
var assets embed.FS

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("Application starting...")

	dataDir := filepath.Join(getAppDataDir(), "FlowCI")
	if err := store.Init(dataDir); err != nil {
		fmt.Printf("Failed to initialize store: %v\n", err)
	}
	fmt.Printf("Data directory: %s\n", dataDir)
}

func getAppDataDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, _ := os.UserHomeDir()
		appData = filepath.Join(home, ".local", "share")
	}
	return appData
}

func (a *App) CheckDocker(ctx context.Context) map[string]interface{} {
	fmt.Println("Checking Docker...")

	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"connected": false,
			"version":   "",
		}
	}

	return map[string]interface{}{
		"connected": true,
		"version":   strings.TrimSpace(string(output)),
	}
}

func (a *App) ListProjects(ctx context.Context) []map[string]interface{} {
	projects, err := store.ListProjects()
	if err != nil {
		fmt.Printf("ListProjects error: %v\n", err)
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(projects))
	for i, p := range projects {
		result[i] = map[string]interface{}{
			"id":         p.ID,
			"name":       p.Name,
			"path":       p.Path,
			"language":   p.Language,
			"created_at": p.CreatedAt.Format(time.RFC3339),
			"updated_at": p.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result
}

func (a *App) ListImages(ctx context.Context) []map[string]interface{} {
	cmd := exec.Command("docker", "images", "--format", "{{.ID}}|{{.Repository}}|{{.Tag}}|{{.Size}}|{{.CreatedAt}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to list docker images: %v\n", err)
		return []map[string]interface{}{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	images := make([]map[string]interface{}, 0)

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			images = append(images, map[string]interface{}{
				"id":         parts[0],
				"repository": parts[1],
				"tag":        parts[2],
				"size":       parts[3],
				"created_at": parts[4],
			})
		}
	}

	return images
}

func (a *App) RemoveImage(ctx context.Context, imageId string) map[string]interface{} {
	cmd := exec.Command("docker", "rmi", "-f", imageId)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		if strings.Contains(outputStr, "No such image") || strings.Contains(outputStr, "not found") {
			return map[string]interface{}{
				"success": false,
				"error":   "镜像不存在",
				"detail":  outputStr,
			}
		}
		if strings.Contains(outputStr, "being used") || strings.Contains(outputStr, "in use") {
			return map[string]interface{}{
				"success": false,
				"error":   "镜像正在使用中，请先停止使用该镜像的容器",
				"detail":  outputStr,
			}
		}
		if strings.Contains(outputStr, "permission denied") || strings.Contains(outputStr, "denied") {
			return map[string]interface{}{
				"success": false,
				"error":   "权限不足，请使用管理员权限运行",
				"detail":  outputStr,
			}
		}
		return map[string]interface{}{
			"success": false,
			"error":   "删除失败",
			"detail":  outputStr,
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": "镜像删除成功",
	}
}

func (a *App) CreateProject(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	input := store.CreateProjectInput{
		Name:     getString(data, "name"),
		Path:     getString(data, "path"),
		Language: getString(data, "language"),
	}

	p, err := store.CreateProject(input)
	if err != nil {
		fmt.Printf("CreateProject error: %v\n", err)
		return nil
	}

	return map[string]interface{}{
		"id":         p.ID,
		"name":       p.Name,
		"path":       p.Path,
		"language":   p.Language,
		"created_at": p.CreatedAt.Format(time.RFC3339),
		"updated_at": p.UpdatedAt.Format(time.RFC3339),
	}
}

func (a *App) DeleteProject(ctx context.Context, projectId string) bool {
	if err := store.DeleteProject(projectId); err != nil {
		fmt.Printf("DeleteProject error: %v\n", err)
		return false
	}
	return true
}

func (a *App) UpdateProject(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	projectId := getString(data, "id")
	if projectId == "" {
		return nil
	}

	input := store.UpdateProjectInput{
		Name:     getString(data, "name"),
		Path:     getString(data, "path"),
		Language: getString(data, "language"),
	}

	p, err := store.UpdateProject(projectId, input)
	if err != nil {
		fmt.Printf("UpdateProject error: %v\n", err)
		return nil
	}

	return map[string]interface{}{
		"id":         p.ID,
		"name":       p.Name,
		"path":       p.Path,
		"language":   p.Language,
		"created_at": p.CreatedAt.Format(time.RFC3339),
		"updated_at": p.UpdatedAt.Format(time.RFC3339),
	}
}

func (a *App) ListPipelines(ctx context.Context, projectId string) []map[string]interface{} {
	pipelines, err := store.ListPipelines(projectId)
	if err != nil {
		fmt.Printf("ListPipelines error: %v\n", err)
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, 0)
	for _, p := range pipelines {
		result = append(result, map[string]interface{}{
			"id":         p.ID,
			"project_id": p.ProjectID,
			"name":       p.Name,
			"steps":      p.Steps,
			"config":     p.Config,
			"created_at": p.CreatedAt.Format(time.RFC3339),
			"updated_at": p.UpdatedAt.Format(time.RFC3339),
		})
	}
	return result
}

func (a *App) CreatePipeline(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	projectId := getString(data, "projectId")
	name := getString(data, "name")
	if projectId == "" || name == "" {
		return map[string]interface{}{"error": "projectId and name are required"}
	}

	var steps []store.PipelineStep
	if stepsData, ok := data["steps"].([]interface{}); ok {
		for _, s := range stepsData {
			if stepMap, ok := s.(map[string]interface{}); ok {
				step := store.PipelineStep{
					Type:   getString(stepMap, "type"),
					Name:   getString(stepMap, "name"),
					Retry:  0,
					OnFail: "stop",
				}
				if v, ok := stepMap["retry"].(float64); ok {
					step.Retry = int(v)
				}
				if v, ok := stepMap["onFail"].(string); ok {
					step.OnFail = v
				}
				if cfg, ok := stepMap["config"].(map[string]interface{}); ok {
					step.Config = cfg
				}
				steps = append(steps, step)
			}
		}
	}

	config := store.PipelineConfig{StopOnFail: true}
	if cfg, ok := data["config"].(map[string]interface{}); ok {
		if v, ok := cfg["stopOnFail"].(bool); ok {
			config.StopOnFail = v
		}
	}

	input := store.CreatePipelineInput{
		ProjectID: projectId,
		Name:     name,
		Steps:    steps,
		Config:   config,
	}

	p, err := store.CreatePipeline(input)
	if err != nil {
		fmt.Printf("CreatePipeline error: %v\n", err)
		return map[string]interface{}{"error": err.Error()}
	}

	return map[string]interface{}{
		"id":         p.ID,
		"project_id": p.ProjectID,
		"name":       p.Name,
		"steps":      p.Steps,
		"config":     p.Config,
		"created_at": p.CreatedAt.Format(time.RFC3339),
	}
}

func (a *App) DeletePipeline(ctx context.Context, pipelineId string) bool {
	if err := store.DeletePipeline(pipelineId); err != nil {
		fmt.Printf("DeletePipeline error: %v\n", err)
		return false
	}
	return true
}

func (a *App) ExportPipelineToYaml(ctx context.Context, pipelineId string) string {
	pipeline, err := store.GetPipeline(pipelineId)
	if err != nil {
		return "# Pipeline not found"
	}

	type YamlStep struct {
		Type   string `yaml:"type"`
		Name   string `yaml:"name"`
		Retry  int    `yaml:"retry,omitempty"`
		OnFail string `yaml:"on_fail,omitempty"`
		Config map[string]interface{} `yaml:"config,omitempty"`
	}

	type YamlConfig struct {
		Parallel   bool `yaml:"parallel,omitempty"`
		StopOnFail bool `yaml:"stop_on_fail"`
	}

	type YamlPipeline struct {
		Name   string     `yaml:"name"`
		Config YamlConfig `yaml:"config"`
		Steps  []YamlStep `yaml:"steps"`
	}

	steps := make([]YamlStep, len(pipeline.Steps))
	for i, s := range pipeline.Steps {
		steps[i] = YamlStep{
			Type:   s.Type,
			Name:   s.Name,
			Retry:  s.Retry,
			OnFail: s.OnFail,
			Config: s.Config,
		}
	}

	yp := YamlPipeline{
		Name: pipeline.Name,
		Config: YamlConfig{
			Parallel:   pipeline.Config.Parallel,
			StopOnFail: pipeline.Config.StopOnFail,
		},
		Steps: steps,
	}

	yamlBytes, err := yaml.Marshal(yp)
	if err != nil {
		return fmt.Sprintf("# Failed to marshal pipeline: %v", err)
	}
	return string(yamlBytes)
}

func (a *App) ImportPipelineFromYaml(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	projectId := getString(data, "projectId")
	yamlContent := getString(data, "yaml")

	if projectId == "" {
		return map[string]interface{}{"error": "projectId is required"}
	}

	if yamlContent == "" {
		return map[string]interface{}{"error": "yaml content is required"}
	}

	type YamlStep struct {
		Type   string `yaml:"type"`
		Name   string `yaml:"name"`
		Retry  int    `yaml:"retry,omitempty"`
		OnFail string `yaml:"on_fail,omitempty"`
		Config map[string]interface{} `yaml:"config,omitempty"`
	}

	type YamlConfig struct {
		Parallel   bool `yaml:"parallel,omitempty"`
		StopOnFail bool `yaml:"stop_on_fail"`
	}

	type YamlPipeline struct {
		Name   string     `yaml:"name"`
		Config YamlConfig `yaml:"config"`
		Steps  []YamlStep `yaml:"steps"`
	}

	var yp YamlPipeline
	if err := yaml.Unmarshal([]byte(yamlContent), &yp); err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("invalid yaml: %v", err)}
	}

	steps := make([]store.PipelineStep, len(yp.Steps))
	validTypes := map[string]bool{"build": true, "push": true, "deploy": true}
	for i, s := range yp.Steps {
		if !validTypes[s.Type] {
			return map[string]interface{}{"error": fmt.Sprintf("invalid step type '%s': must be one of build, push, deploy", s.Type)}
		}
		steps[i] = store.PipelineStep{
			Type:   s.Type,
			Name:   s.Name,
			Retry:  s.Retry,
			OnFail: s.OnFail,
			Config: s.Config,
		}
	}

	input := store.CreatePipelineInput{
		ProjectID: projectId,
		Name:     yp.Name,
		Steps:    steps,
		Config: store.PipelineConfig{
			Parallel:   yp.Config.Parallel,
			StopOnFail: yp.Config.StopOnFail,
		},
	}

	pipeline, err := store.CreatePipeline(input)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	return map[string]interface{}{
		"id":         pipeline.ID,
		"name":       pipeline.Name,
		"steps":      pipeline.Steps,
		"created_at": pipeline.CreatedAt.Format(time.RFC3339),
	}
}

func (a *App) ExecutePipeline(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	pipelineId := getString(data, "pipelineId")
	projectId := getString(data, "projectId")

	pipeline, err := store.GetPipeline(pipelineId)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}

	logs := []map[string]interface{}{}
	allSuccess := true

	for _, step := range pipeline.Steps {
		stepLog := map[string]interface{}{
			"step":    step.Name,
			"type":    step.Type,
			"status":  "pending",
			"message": "",
		}

		var stepErr error
		for retry := 0; retry <= step.Retry; retry++ {
			if retry > 0 {
				fmt.Printf("Retrying step %s (attempt %d)\n", step.Name, retry)
			}

			switch step.Type {
			case "build":
				tag := "latest"
				if t, ok := step.Config["tag"].(string); ok {
					tag = t
				}
				contextPath := "."
				if cp, ok := step.Config["contextPath"].(string); ok {
					contextPath = cp
				}
				result := buildImage(projectId, tag, contextPath, false, false)
				if !result["success"].(bool) {
					stepErr = fmt.Errorf("%v", result["error"])
				} else {
					stepErr = nil
					stepLog["message"] = fmt.Sprintf("Built: %s:%s", result["image_name"], result["image_tag"])
				}

			case "push":
				imageName := ""
				if img, ok := step.Config["imageName"].(string); ok {
					imageName = img
				}
				registry := ""
				if reg, ok := step.Config["registry"].(string); ok {
					registry = reg
				}
				username := ""
				if u, ok := step.Config["username"].(string); ok {
					username = u
				}
				password := ""
				if pwd, ok := step.Config["password"].(string); ok {
					password = pwd
				}
				result := pushImageWithCreds(imageName, registry, username, password)
				if !result["success"].(bool) {
					stepErr = fmt.Errorf("%v", result["error"])
				} else {
					stepErr = nil
					stepLog["message"] = "Image pushed successfully"
				}

			case "deploy":
				imageName := ""
				if img, ok := step.Config["imageName"].(string); ok {
					imageName = img
				}
				name := ""
				if n, ok := step.Config["name"].(string); ok {
					name = n
				}
				hostPort := ""
				if hp, ok := step.Config["hostPort"].(string); ok {
					hostPort = hp
				}
				containerPort := ""
				if cp, ok := step.Config["containerPort"].(string); ok {
					containerPort = cp
				}
				restartPolicy := "unless-stopped"
				if rp, ok := step.Config["restartPolicy"].(string); ok {
					restartPolicy = rp
				}
				result := deployContainer(imageName, name, hostPort, containerPort, restartPolicy, "")
				if !result["success"].(bool) {
					stepErr = fmt.Errorf("%v", result["error"])
				} else {
					stepErr = nil
					stepLog["message"] = "Container deployed successfully"
				}
			}

			if stepErr == nil {
				break
			}
		}

		if stepErr != nil {
			stepLog["status"] = "failed"
			stepLog["error"] = stepErr.Error()
			logs = append(logs, stepLog)
			if step.OnFail == "stop" || pipeline.Config.StopOnFail {
				allSuccess = false
				break
			}
		} else {
			stepLog["status"] = "success"
			logs = append(logs, stepLog)
		}
	}

	return map[string]interface{}{
		"success": allSuccess,
		"logs":    logs,
		"message": map[bool]string{true: "Pipeline executed successfully", false: "Pipeline failed"}[allSuccess],
	}, nil
}

func buildImage(projectId, tag, contextPath string, noCache, pullLatest bool) map[string]interface{} {
	if projectId == "" {
		return map[string]interface{}{"success": false, "error": "projectId is required"}
	}

	record, err := store.CreateBuildRecord(projectId, tag, "latest")
	if err != nil {
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("create build record: %v", err)}
	}

	args := []string{"buildx", "build", "-t", tag, "--progress=plain"}
	if noCache {
		args = append(args, "--no-cache")
	}
	if pullLatest {
		args = append(args, "--pull")
	}
	args = append(args, contextPath)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	logStr := string(output)

	if err != nil {
		store.FinishBuildRecord(record.ID, "failed", logStr)
		parts := strings.SplitN(tag, ":", 2)
		imageName := parts[0]
		imageTag := "latest"
		if len(parts) > 1 {
			imageTag = parts[1]
		}
		return map[string]interface{}{
			"success":    false,
			"image_name": imageName,
			"image_tag":  imageTag,
			"log":        logStr,
			"error":      err.Error(),
		}
	}

	store.FinishBuildRecord(record.ID, "success", logStr)
	parts := strings.SplitN(tag, ":", 2)
	imageName := parts[0]
	imageTag := "latest"
	if len(parts) > 1 {
		imageTag = parts[1]
	}

	return map[string]interface{}{
		"success":    true,
		"image_name": imageName,
		"image_tag":  imageTag,
		"log":        logStr,
	}
}

func pushImage(imageName, registry string) map[string]interface{} {
	return pushImageWithCreds(imageName, registry, "", "")
}

func pushImageWithCreds(imageName, registry, username, password string) map[string]interface{} {
	targetImage := imageName

	if registry != "" && registry != "docker.io" {
		if username != "" && password != "" {
			loginCmd := exec.Command("docker", "login", registry, "-u", username, "--password-stdin")
			loginCmd.Stdin = strings.NewReader(password)
			if out, err := loginCmd.CombinedOutput(); err != nil {
				return map[string]interface{}{"success": false, "error": fmt.Sprintf("login failed: %s", string(out))}
			}
		}
		targetImage = registry + "/" + imageName
		cmd := exec.Command("docker", "tag", imageName, targetImage)
		if err := cmd.Run(); err != nil {
			return map[string]interface{}{"success": false, "error": fmt.Sprintf("tag failed: %v", err)}
		}
		imageName = targetImage
	} else if registry == "docker.io" && username != "" && password != "" {
		loginCmd := exec.Command("docker", "login", "-u", username, "--password-stdin")
		loginCmd.Stdin = strings.NewReader(password)
		if out, err := loginCmd.CombinedOutput(); err != nil {
			return map[string]interface{}{"success": false, "error": fmt.Sprintf("docker hub login failed: %s", string(out))}
		}
	}

	cmd := exec.Command("docker", "push", targetImage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("push failed: %s", string(output))}
	}

	return map[string]interface{}{"success": true, "message": "Image pushed successfully"}
}

func deployContainer(imageName, name, hostPort, containerPort, restartPolicy, envStr string) map[string]interface{} {
	if restartPolicy == "" {
		restartPolicy = "unless-stopped"
	}

	args := []string{"run", "-d", "--name", name, "--restart", restartPolicy}

	if hostPort != "" && containerPort != "" {
		args = append(args, "-p", hostPort+":"+containerPort)
	}

	if envStr != "" {
		for _, line := range strings.Split(envStr, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				args = append(args, "-e", line)
			}
		}
	}

	args = append(args, imageName)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": string(output),
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"id":      strings.TrimSpace(string(output)),
		"message": "Container deployed successfully",
	}
}

func (a *App) ListBuildRecords(ctx context.Context, projectId string) []map[string]interface{} {
	records, err := store.ListBuildRecords(projectId)
	if err != nil {
		fmt.Printf("ListBuildRecords error: %v\n", err)
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(records))
	for i, r := range records {
		m := map[string]interface{}{
			"id":         r.ID,
			"project_id": r.ProjectID,
			"image_name": r.ImageName,
			"image_tag":  r.ImageTag,
			"status":     r.Status,
			"started_at": r.StartedAt.Format(time.RFC3339),
		}
		if r.FinishedAt != nil {
			m["finished_at"] = r.FinishedAt.Format(time.RFC3339)
		}
		result[i] = m
	}
	return result
}

func (a *App) GetBuildRecord(ctx context.Context, recordId string) map[string]interface{} {
	r, err := store.GetBuildRecord(recordId)
	if err != nil {
		fmt.Printf("GetBuildRecord error: %v\n", err)
		return nil
	}

	result := map[string]interface{}{
		"id":         r.ID,
		"project_id": r.ProjectID,
		"image_name": r.ImageName,
		"image_tag":  r.ImageTag,
		"status":     r.Status,
		"log":        r.Log,
		"started_at": r.StartedAt.Format(time.RFC3339),
	}
	if r.FinishedAt != nil {
		result["finished_at"] = r.FinishedAt.Format(time.RFC3339)
	}
	return result
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (a *App) BuildImage(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	contextPath := getString(data, "contextPath")
	tag := getString(data, "tag")
	noCache := false
	pullLatest := false
	if v, ok := data["noCache"]; ok {
		noCache, _ = v.(bool)
	}
	if v, ok := data["pullLatest"]; ok {
		pullLatest, _ = v.(bool)
	}

	projectID := getString(data, "projectId")
	if projectID == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "projectId is required",
		}, nil
	}

	if _, err := store.GetProject(projectID); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("project not found: %s", projectID),
		}, nil
	}

	record, err := store.CreateBuildRecord(projectID, tag, "latest")
	if err != nil {
		return nil, fmt.Errorf("create build record: %w", err)
	}

	args := []string{"buildx", "build", "-t", tag, "--progress=plain"}
	if noCache {
		args = append(args, "--no-cache")
	}
	if pullLatest {
		args = append(args, "--pull")
	}
	args = append(args, contextPath)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	logStr := string(output)

	if err != nil {
		store.FinishBuildRecord(record.ID, "failed", logStr)
		parts := strings.SplitN(tag, ":", 2)
		imageName := parts[0]
		imageTag := "latest"
		if len(parts) > 1 {
			imageTag = parts[1]
		}
		return map[string]interface{}{
			"success":    false,
			"image_name": imageName,
			"image_tag":  imageTag,
			"log":        logStr,
			"error":      err.Error(),
		}, nil
	}

	store.FinishBuildRecord(record.ID, "success", logStr)
	parts := strings.SplitN(tag, ":", 2)
	imageName := parts[0]
	imageTag := "latest"
	if len(parts) > 1 {
		imageTag = parts[1]
	}

	return map[string]interface{}{
		"success":    true,
		"image_name": imageName,
		"image_tag":  imageTag,
		"log":        logStr,
	}, nil
}

func (a *App) ListContainers(ctx context.Context) []map[string]interface{} {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.ID}}|{{.Names}}|{{.Image}}|{{.State}}|{{.Status}}|{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return []map[string]interface{}{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	containers := make([]map[string]interface{}, 0)

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			containers = append(containers, map[string]interface{}{
				"id":     parts[0],
				"names":  []string{parts[1]},
				"image":  parts[2],
				"state":  parts[3],
				"status": parts[4],
				"ports":  parts[5],
			})
		}
	}

	return containers
}

func (a *App) StartContainer(ctx context.Context, containerId string) bool {
	fmt.Printf("Starting container: %s\n", containerId)
	cmd := exec.Command("docker", "start", containerId)
	err := cmd.Run()
	return err == nil
}

func (a *App) StopContainer(ctx context.Context, containerId string) bool {
	fmt.Printf("Stopping container: %s\n", containerId)
	cmd := exec.Command("docker", "stop", containerId)
	err := cmd.Run()
	return err == nil
}

func (a *App) RemoveContainer(ctx context.Context, containerId string) bool {
	fmt.Printf("Removing container: %s\n", containerId)
	cmd := exec.Command("docker", "rm", "-f", containerId)
	err := cmd.Run()
	return err == nil
}

func (a *App) GetContainerLogs(ctx context.Context, containerId string, tailLines int) string {
	if tailLines <= 0 {
		tailLines = 100
	}
	cmd := exec.Command("docker", "logs", "--tail", fmt.Sprintf("%d", tailLines), containerId)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("获取日志失败: %v", err)
	}
	return string(output)
}

func (a *App) DeployContainer(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	image := getString(data, "image")
	name := getString(data, "name")
	hostPort := getString(data, "hostPort")
	containerPort := getString(data, "containerPort")
	restartPolicy := getString(data, "restartPolicy")
	if restartPolicy == "" {
		restartPolicy = "unless-stopped"
	}

	args := []string{"run", "-d", "--name", name, "--restart", restartPolicy}

	if hostPort != "" && containerPort != "" {
		args = append(args, "-p", hostPort+":"+containerPort)
	}

	envStr := getString(data, "env")
	if envStr != "" {
		for _, line := range strings.Split(envStr, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				args = append(args, "-e", line)
			}
		}
	}

	args = append(args, image)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": string(output),
			"error":   err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"id":      strings.TrimSpace(string(output)),
		"message": "Container deployed successfully",
	}, nil
}

func (a *App) GenerateCompose(ctx context.Context, data map[string]interface{}) string {
	image := getString(data, "image")
	name := getString(data, "name")
	hostPort := getString(data, "hostPort")
	containerPort := getString(data, "containerPort")
	restartPolicy := getString(data, "restartPolicy")
	if restartPolicy == "" {
		restartPolicy = "unless-stopped"
	}
	envStr := getString(data, "env")

	compose := fmt.Sprintf(`version: '3.8'
services:
  %s:
    image: %s
    container_name: %s
    restart: %s
`, name, image, name, restartPolicy)

	if hostPort != "" && containerPort != "" {
		compose += fmt.Sprintf(`    ports:
      - "%s:%s"
`, hostPort, containerPort)
	}

	if envStr != "" {
		compose += `    environment:
`
		for _, line := range strings.Split(envStr, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				compose += fmt.Sprintf(`      - "%s"
`, line)
			}
		}
	}

	return compose
}

func (a *App) DeployWithCompose(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	composeContent := getString(data, "compose")
	workDir := getString(data, "workdir")

	if workDir == "" {
		workDir = "."
	}

	tmpFile := filepath.Join(workDir, "docker-compose.yml")
	if err := os.WriteFile(tmpFile, []byte(composeContent), 0644); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to write docker-compose.yml: " + err.Error(),
		}, nil
	}

	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": string(output),
			"error":   err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"message": "Deployed with docker-compose successfully",
		"output":  string(output),
	}, nil
}

func (a *App) GetSupportedLanguages(ctx context.Context) []map[string]string {
	return []map[string]string{
		{"language": "nodejs", "display_name": "🟢 Node.js"},
		{"language": "python", "display_name": "🐍 Python"},
		{"language": "go", "display_name": "🔵 Go"},
		{"language": "java-maven", "display_name": "☕ Java (Maven)"},
		{"language": "java-gradle", "display_name": "☕ Java (Gradle)"},
		{"language": "php", "display_name": "🐘 PHP"},
		{"language": "ruby", "display_name": "💎 Ruby"},
		{"language": "dotnet", "display_name": "🔷 .NET"},
		{"language": "rust", "display_name": "🦀 Rust"},
		{"language": "c", "display_name": "⚙️ C/C++"},
	}
}

func (a *App) GetSettings(ctx context.Context) map[string]string {
	settings, err := store.GetSettings()
	if err != nil {
		fmt.Printf("GetSettings error: %v\n", err)
		return map[string]string{}
	}
	return settings
}

func (a *App) SaveSettings(ctx context.Context, data map[string]interface{}) bool {
	for key, val := range data {
		if str, ok := val.(string); ok {
			if err := store.SaveSettings(key, str); err != nil {
				fmt.Printf("SaveSettings error: %v\n", err)
				return false
			}
		}
	}
	return true
}

func (a *App) GenerateDockerfile(ctx context.Context, language string) string {
	templates := map[string]string{
		"nodejs": `FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]`,
		"python": `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["python", "main.py"]`,
		"go": `FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]`,
		"java-maven": `FROM maven:3.9-eclipse-temurin-21 AS builder
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline
COPY . .
RUN mvn package -DskipTests

FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=builder /app/target/*.jar app.jar
EXPOSE 8080
CMD ["java", "-jar", "app.jar"]`,
		"java-gradle": `FROM gradle:8-jdk21 AS builder
WORKDIR /app
COPY build.gradle* settings.gradle* ./
RUN gradle dependencies --no-daemon
COPY . .
RUN gradle build -x test --no-daemon

FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=builder /app/build/libs/*.jar app.jar
EXPOSE 8080
CMD ["java", "-jar", "app.jar"]`,
		"php": `FROM php:8.2-apache
WORKDIR /var/www/html
COPY . .
RUN docker-php-ext-install pdo pdo_mysql
EXPOSE 80
CMD ["apache2-foreground"]`,
		"ruby": `FROM ruby:3.2-alpine
WORKDIR /app
COPY Gemfile Gemfile.lock ./
RUN bundle install
COPY . .
EXPOSE 4567
CMD ["ruby", "app.rb"]`,
		"dotnet": `FROM mcr.microsoft.com/dotnet/sdk:8.0 AS builder
WORKDIR /app
COPY *.csproj .
RUN dotnet restore
COPY . .
RUN dotnet publish -c Release -o out

FROM mcr.microsoft.com/dotnet/aspnet:8.0
WORKDIR /app
COPY --from=builder /app/out .
EXPOSE 8080
CMD ["dotnet", "app.dll"]`,
		"rust": `FROM rust:1.75-alpine AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release 2>/dev/null || true
COPY . .
RUN cargo build --release

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/target/release/* .
EXPOSE 8080
CMD ["./app"]`,
		"c": `FROM gcc:13-bookworm AS builder
WORKDIR /app
COPY . .
RUN gcc -O2 -o app main.c -lpthread

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]`,
	}
	if tmpl, ok := templates[language]; ok {
		return tmpl
	}
	return "# Dockerfile for " + language
}

func (a *App) PushImage(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	image := getString(data, "image")
	if image == "" {
		return map[string]interface{}{"success": false, "error": "image is required"}, nil
	}

	registry := getString(data, "registry")
	username := getString(data, "username")
	password := getString(data, "password")
	targetImage := image

	if registry != "" && registry != "docker.io" {
		if username != "" && password != "" {
			loginCmd := exec.Command("docker", "login", registry, "-u", username, "--password-stdin")
			loginCmd.Stdin = strings.NewReader(password)
			if err := loginCmd.Run(); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("login failed: %v", err),
				}, nil
			}
		}
		targetImage = registry + "/" + image
		cmd := exec.Command("docker", "tag", image, targetImage)
		if err := cmd.Run(); err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("tag failed: %v", err),
			}, nil
		}
	} else if registry == "docker.io" && username != "" && password != "" {
		loginCmd := exec.Command("docker", "login", "-u", username, "--password-stdin")
		loginCmd.Stdin = strings.NewReader(password)
		if err := loginCmd.Run(); err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("docker hub login failed: %v", err),
			}, nil
		}
	}

	cmd := exec.Command("docker", "push", targetImage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   string(output),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"log":     string(output),
	}, nil
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "FlowCI",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		fmt.Println("Error:", err)
	}
}
