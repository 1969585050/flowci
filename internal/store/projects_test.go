package store

import (
	"errors"
	"testing"
	"time"
)

func TestCreateAndGetProject(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	p, err := CreateProject(CreateProjectInput{
		Name: "demo", Path: "/tmp/demo", Language: "go",
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if p.Name != "demo" || p.Path != "/tmp/demo" || p.Language != "go" {
		t.Errorf("unexpected fields: %+v", p)
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		t.Error("timestamps should be set")
	}

	got, err := GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.ID != p.ID || got.Name != p.Name {
		t.Errorf("round-trip mismatch: %+v vs %+v", got, p)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	_, err := GetProject("nonexistent-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateProject(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	created, _ := CreateProject(CreateProjectInput{Name: "old", Path: "/a", Language: "go"})

	// 让 UpdatedAt 变化
	time.Sleep(10 * time.Millisecond)

	updated, err := UpdateProject(created.ID, UpdateProjectInput{
		Name: "new", Path: "/b", Language: "python",
	})
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Name != "new" || updated.Path != "/b" || updated.Language != "python" {
		t.Errorf("update fields not applied: %+v", updated)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt.Time) {
		t.Errorf("UpdatedAt not advanced: %v vs %v", updated.UpdatedAt, created.UpdatedAt)
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	_, err := UpdateProject("nonexistent", UpdateProjectInput{Name: "x"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteProject(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	p, _ := CreateProject(CreateProjectInput{Name: "x", Path: "/x", Language: "go"})

	if err := DeleteProject(p.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	_, err := GetProject(p.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("deleted project should be gone, got %v", err)
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	err := DeleteProject("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListProjects_OrderByUpdatedDesc(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	p1, _ := CreateProject(CreateProjectInput{Name: "first", Path: "/1", Language: "go"})
	time.Sleep(10 * time.Millisecond)
	p2, _ := CreateProject(CreateProjectInput{Name: "second", Path: "/2", Language: "go"})
	time.Sleep(10 * time.Millisecond)
	// 更新 p1 让它变成最新
	_, _ = UpdateProject(p1.ID, UpdateProjectInput{Name: "first-updated", Path: "/1", Language: "go"})

	list, err := ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(list))
	}
	// p1 刚被更新，应该排第一
	if list[0].ID != p1.ID {
		t.Errorf("expected p1 first (just updated), got %q", list[0].ID)
	}
	if list[1].ID != p2.ID {
		t.Errorf("expected p2 second, got %q", list[1].ID)
	}
}

func TestListProjects_Empty(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	list, err := ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}
