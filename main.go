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

	args := []string{"build", "-t", tag}
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

	cmd := exec.Command("docker", "push", image)
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
