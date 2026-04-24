package docker

import (
	"context"
	"errors"
	"strings"
)

// Image 表示一条 docker images 列表项。
type Image struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Size       string `json:"size"`
	CreatedAt  string `json:"createdAt"`
}

// 镜像删除错误的业务分类。
var (
	ErrImageNotFound = errors.New("image not found")
	ErrImageInUse    = errors.New("image in use")
)

// ListImages 返回本机所有 docker 镜像。
// 格式解析靠 `docker images --format {{.X}}|...|` 约定，字段内不含 | 是 docker 自身保证。
func ListImages(ctx context.Context) ([]Image, error) {
	out, err := run(ctx, TimeoutQuery, "images",
		"--format", "{{.ID}}|{{.Repository}}|{{.Tag}}|{{.Size}}|{{.CreatedAt}}")
	if err != nil {
		return nil, err
	}
	return parseImages(string(out)), nil
}

func parseImages(s string) []Image {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return []Image{}
	}
	lines := strings.Split(trimmed, "\n")
	images := make([]Image, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 5 {
			continue
		}
		images = append(images, Image{
			ID:         parts[0],
			Repository: parts[1],
			Tag:        parts[2],
			Size:       parts[3],
			CreatedAt:  parts[4],
		})
	}
	return images
}

// RemoveImage 强制删除指定 ID 的镜像。
// 返回值可能是 ErrImageNotFound / ErrImageInUse，或原始 run 错误。
func RemoveImage(ctx context.Context, id string) error {
	out, err := run(ctx, TimeoutLifecycle, "rmi", "-f", id)
	if err == nil {
		return nil
	}
	s := strings.ToLower(string(out))
	switch {
	case strings.Contains(s, "no such image"), strings.Contains(s, "not found"):
		return ErrImageNotFound
	case strings.Contains(s, "being used"), strings.Contains(s, "in use"):
		return ErrImageInUse
	}
	return err
}
