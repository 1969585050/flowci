package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/flowci/flowci/pkg/docker"
)

type Builder struct {
	docker *docker.Client
}

func NewBuilder(dockerClient *docker.Client) *Builder {
	return &Builder{docker: dockerClient}
}

type BuildConfig struct {
	ProjectID       string
	Language        Language
	ContextPath     string
	DockerfilePath  string
	ImageTags       []string
	BuildArgs       map[string]string
	NoCache         bool
	PullBaseImage   bool
}

type Language string

const (
	LangJavaMaven  Language = "java-maven"
	LangJavaGradle Language = "java-gradle"
	LangNodeJS     Language = "nodejs"
	LangPython     Language = "python"
	LangGo         Language = "go"
	LangPHP        Language = "php"
	LangRuby       Language = "ruby"
	LangDotnet     Language = "dotnet"
	LangCustom     Language = "custom"
)

type BuildResult struct {
	ImageID   string
	ImageTags []string
	Size      int64
	Duration  time.Duration
}

func (b *Builder) Build(ctx context.Context, cfg *BuildConfig) (*BuildResult, error) {
	dockerCli := b.docker.GetCLI()

	tarCtx, err := createBuildContext(cfg.ContextPath, cfg.DockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create build context: %w", err)
	}
	defer tarCtx.Close()

	buildArgs := make(map[string]*string)
	for k, v := range cfg.BuildArgs {
		v := v
		buildArgs[k] = &v
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       cfg.ImageTags,
		BuildArgs:  buildArgs,
		NoCache:    cfg.NoCache,
		Remove:     true,
		PullParent: cfg.PullBaseImage,
	}

	resp, err := dockerCli.ImageBuild(ctx, tarCtx, opts)
	if err != nil {
		return nil, fmt.Errorf("image build failed: %w", err)
	}
	defer resp.Body.Close()

	imageID := ""
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read build response: %w", err)
	}

	imageID = extractImageID(resp.OSType, resp.ID)

	return &BuildResult{
		ImageID:   imageID,
		ImageTags: cfg.ImageTags,
		Duration:  time.Since(time.Now()),
	}, nil
}

func (b *Builder) GenerateDockerfile(lang Language) string {
	templates := map[Language]string{
		LangJavaMaven: `FROM maven:3.9-eclipse-temurin-17 AS build
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline -B
COPY src ./src
RUN mvn clean package -DskipTests

FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`,

		LangJavaGradle: `FROM gradle:8.5-jdk17 AS build
WORKDIR /app
COPY build.gradle settings.gradle ./
RUN gradle dependencies --no-daemon
COPY src ./src
RUN gradle build -x test --no-daemon

FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY --from=build /app/build/libs/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`,

		LangNodeJS: `FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build || echo "No build step"

FROM node:20-alpine
WORKDIR /app
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
EXPOSE 3000
CMD ["node", "dist/index.js"]`,

		LangPython: `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["python", "main.py"]`,

		LangGo: `FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/app .
EXPOSE 8080
CMD ["./app"]`,

		LangPHP: `FROM php:8.2-cli-alpine
WORKDIR /app
COPY composer.* ./
RUN composer install --no-dev --no-scripts
COPY . .
EXPOSE 8080
CMD ["php", "-S", "0.0.0.0:8080"]`,

		LangRuby: `FROM ruby:3.2-alpine
WORKDIR /app
COPY Gemfile Gemfile.lock ./
RUN bundle install
COPY . .
EXPOSE 3000
CMD ["ruby", "main.rb"]`,

		LangDotnet: `FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src
COPY . .
RUN dotnet restore
RUN dotnet build -c Release -o /app

FROM mcr.microsoft.com/dotnet/aspnet:8.0
WORKDIR /app
COPY --from=build /app .
EXPOSE 8080
CMD ["dotnet", "app.dll"]`,
	}

	if tmpl, ok := templates[lang]; ok {
		return tmpl
	}
	return templates[LangCustom]
}

func createBuildContext(contextPath, dockerfilePath string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte{})), nil
}

func extractImageID(osType, id string) string {
	if id == "" {
		return "latest"
	}
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
