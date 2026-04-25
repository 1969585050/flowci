package handler

import (
	"log/slog"
	"time"

	"flowci/internal/store"
)

// GetDashboardStats 一次拉全部首页概览数据。
//
// 失败容忍：单项数据失败不阻塞其它项（用 slog 记录），让前端尽量看到能看的部分。
// docker 连通失败时容器/镜像数为 0；DB 查询失败时对应字段为 0。
func (a *App) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 项目数 + Git 项目数
	if projects, err := store.ListProjects(); err == nil {
		stats.Projects = len(projects)
		for _, p := range projects {
			if p.RepoURL != "" {
				stats.GitProjects++
			}
		}
	} else {
		slog.Warn("dashboard: list projects failed", "err", err)
	}

	// 流水线数
	if n, err := store.CountPipelines(); err == nil {
		stats.Pipelines = n
	} else {
		slog.Warn("dashboard: count pipelines failed", "err", err)
	}

	// 最近 10 条全局构建
	if recent, err := store.RecentBuildsAcrossProjects(10); err == nil {
		stats.RecentBuilds = recent
	} else {
		slog.Warn("dashboard: recent builds failed", "err", err)
		stats.RecentBuilds = []store.BuildRecord{}
	}

	// 24h 构建摘要
	since := time.Now().Add(-24 * time.Hour).UTC()
	if s, f, b, err := store.CountBuildsByStatusSince(since); err == nil {
		stats.BuildSummary = BuildSummaryStats{Success: s, Failed: f, Building: b}
	} else {
		slog.Warn("dashboard: build summary failed", "err", err)
	}

	// docker 连通 + 容器/镜像
	stats.Docker = a.docker.Check(a.ctx)
	if stats.Docker.Connected {
		if containers, err := a.docker.ListContainers(a.ctx); err == nil {
			stats.Containers.Total = len(containers)
			for _, c := range containers {
				if c.State == "running" {
					stats.Containers.Running++
				} else {
					stats.Containers.Stopped++
				}
			}
		}
		if images, err := a.docker.ListImages(a.ctx); err == nil {
			stats.Images = len(images)
		}
	}

	return stats, nil
}
