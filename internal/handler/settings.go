package handler

import (
	"fmt"

	"flowci/internal/docker"
	"flowci/internal/store"
)

// GetSettings 返回所有设置键值对（key-value 皆 string）。
func (a *App) GetSettings() (map[string]string, error) {
	return store.GetSettings()
}

// SaveSettings 批量写入设置；保存 dockerHost 时同步 docker 包的运行时变量。
// TODO(phase-3)：包在事务里保证原子性；敏感 key（password/token）走 keyring。
func (a *App) SaveSettings(req *SaveSettingsRequest) error {
	if req == nil || req.Settings == nil {
		return fmt.Errorf("%w: settings required", ErrBadRequest)
	}
	for k, v := range req.Settings {
		if err := store.SaveSettings(k, v); err != nil {
			return err
		}
		if k == "dockerHost" {
			docker.SetDockerHost(v)
		}
	}
	return nil
}

// DetectDockerEnv 一次性探测目标 Docker 环境（version/buildx/compose）。
// req.Host 为空时用当前已生效的 DOCKER_HOST（可在保存前用输入框的值试探）。
func (a *App) DetectDockerEnv(req *DetectDockerEnvRequest) docker.EnvReport {
	host := ""
	if req != nil {
		host = req.Host
	}
	return docker.DetectEnv(a.ctx, host)
}

// GetSupportedLanguages 返回前端 Dockerfile 生成器支持的语言列表。
func (a *App) GetSupportedLanguages() []Language {
	return []Language{
		{Language: "nodejs", DisplayName: "🟢 Node.js"},
		{Language: "python", DisplayName: "🐍 Python"},
		{Language: "go", DisplayName: "🔵 Go"},
		{Language: "java-maven", DisplayName: "☕ Java (Maven)"},
		{Language: "java-gradle", DisplayName: "☕ Java (Gradle)"},
		{Language: "php", DisplayName: "🐘 PHP"},
		{Language: "ruby", DisplayName: "💎 Ruby"},
		{Language: "dotnet", DisplayName: "🔷 .NET"},
		{Language: "rust", DisplayName: "🦀 Rust"},
		{Language: "c", DisplayName: "⚙️ C/C++"},
	}
}

// GenerateDockerfile 返回指定语言的 Dockerfile 模板。
// 不支持的语言返回 ErrUnsupportedLang。
func (a *App) GenerateDockerfile(language string) (string, error) {
	tmpl, ok := dockerfileTemplates[language]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedLang, language)
	}
	return tmpl, nil
}

// dockerfileTemplates 按语言归档模板；保持和原 main.go 完全一致的文本以免前端回归。
// TODO(future)：外部化为 embed fs，支持用户自定义覆盖。
var dockerfileTemplates = map[string]string{
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
