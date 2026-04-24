package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("Application starting...")
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
	return []map[string]interface{}{
		{
			"id":         "proj-1",
			"name":       "示例项目 (Node.js)",
			"path":       "/workspace/my-app",
			"language":   "nodejs",
			"created_at": time.Now().Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		},
		{
			"id":         "proj-2",
			"name":       "API 服务 (Go)",
			"path":       "/workspace/api-service",
			"language":   "go",
			"created_at": time.Now().Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		},
	}
}

func (a *App) CreateProject(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	fmt.Printf("Creating project: %v\n", data)
	return map[string]interface{}{
		"id":         "new-proj",
		"name":       data["name"],
		"path":       data["path"],
		"language":   data["language"],
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}
}

func (a *App) DeleteProject(ctx context.Context, projectId string) bool {
	fmt.Printf("Deleting project: %s\n", projectId)
	return true
}

func (a *App) BuildImage(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	fmt.Printf("Building image: %v\n", data)
	return map[string]interface{}{
		"id":          "build-123",
		"image_id":    "sha256:abc123",
		"tags":        []string{"myapp:latest"},
		"size":        125000000,
		"duration_ms": 45000,
		"status":      "success",
		"started_at":  time.Now().Add(-45 * time.Second).Format(time.RFC3339),
		"finished_at": time.Now().Format(time.RFC3339),
	}
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

func (a *App) DeployContainer(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	fmt.Printf("Deploying container: %v\n", data)
	return map[string]interface{}{
		"id":      "deploy-123",
		"status":  "deployed",
		"message": "Container deployed successfully",
	}
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
		"go": `FROM golang:1.21-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
EXPOSE 8080
CMD ["./main"]`,
	}
	if tmpl, ok := templates[language]; ok {
		return tmpl
	}
	return "# Dockerfile for " + language
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "FlowCI",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: nil,
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
