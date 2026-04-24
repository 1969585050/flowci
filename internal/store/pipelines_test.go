package store

import (
	"errors"
	"testing"
)

// createTestProject 为 pipeline 测试快速准备一个项目，返回 project ID。
func createTestProject(t *testing.T) string {
	t.Helper()
	p, err := CreateProject(CreateProjectInput{Name: "test", Path: "/tmp/t", Language: "go"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	return p.ID
}

func TestCreateAndGetPipeline(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	p, err := CreatePipeline(CreatePipelineInput{
		ProjectID: pid,
		Name:      "build-and-push",
		Steps: []PipelineStep{
			{Type: "build", Name: "b1", Config: map[string]interface{}{"tag": "v1"}, Retry: 0, OnFail: "stop"},
			{Type: "push", Name: "p1", Config: map[string]interface{}{}, Retry: 2, OnFail: "continue"},
		},
		Config: PipelineConfig{Parallel: false, StopOnFail: true},
	})
	if err != nil {
		t.Fatalf("CreatePipeline: %v", err)
	}
	if p.ID == "" {
		t.Fatal("ID empty")
	}

	got, err := GetPipeline(p.ID)
	if err != nil {
		t.Fatalf("GetPipeline: %v", err)
	}
	if got.Name != "build-and-push" || len(got.Steps) != 2 {
		t.Errorf("round-trip lost fields: %+v", got)
	}
	if got.Steps[0].Type != "build" || got.Steps[1].OnFail != "continue" {
		t.Errorf("step fields wrong: %+v", got.Steps)
	}
	if !got.Config.StopOnFail {
		t.Errorf("Config.StopOnFail = false, want true")
	}
}

func TestGetPipeline_NotFound(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	_, err := GetPipeline("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdatePipeline(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	created, _ := CreatePipeline(CreatePipelineInput{
		ProjectID: pid, Name: "old",
		Steps:  []PipelineStep{{Type: "build", Name: "s1"}},
		Config: PipelineConfig{StopOnFail: true},
	})

	updated, err := UpdatePipeline(created.ID, UpdatePipelineInput{
		Name: "new",
		Steps: []PipelineStep{
			{Type: "deploy", Name: "d1", Retry: 1},
		},
		Config: PipelineConfig{Parallel: true, StopOnFail: false},
	})
	if err != nil {
		t.Fatalf("UpdatePipeline: %v", err)
	}
	if updated.Name != "new" || len(updated.Steps) != 1 || updated.Steps[0].Type != "deploy" {
		t.Errorf("fields not applied: %+v", updated)
	}
	if !updated.Config.Parallel {
		t.Errorf("Config.Parallel not propagated")
	}
}

func TestUpdatePipeline_NotFound(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	_, err := UpdatePipeline("nonexistent", UpdatePipelineInput{Name: "x"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeletePipeline(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	p, _ := CreatePipeline(CreatePipelineInput{
		ProjectID: pid, Name: "doomed",
		Steps:  []PipelineStep{{Type: "build", Name: "s1"}},
		Config: PipelineConfig{StopOnFail: true},
	})

	if err := DeletePipeline(p.ID); err != nil {
		t.Fatalf("DeletePipeline: %v", err)
	}

	_, err := GetPipeline(p.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("deleted pipeline should be gone, got %v", err)
	}
}

func TestDeletePipeline_NotFound(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	err := DeletePipeline("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListPipelines_ByProject(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	p1, _ := CreateProject(CreateProjectInput{Name: "p1", Path: "/1", Language: "go"})
	p2, _ := CreateProject(CreateProjectInput{Name: "p2", Path: "/2", Language: "go"})

	for _, n := range []string{"a", "b"} {
		_, _ = CreatePipeline(CreatePipelineInput{
			ProjectID: p1.ID, Name: n,
			Steps: []PipelineStep{{Type: "build", Name: "s"}}, Config: PipelineConfig{StopOnFail: true},
		})
	}
	_, _ = CreatePipeline(CreatePipelineInput{
		ProjectID: p2.ID, Name: "other",
		Steps: []PipelineStep{{Type: "push", Name: "s"}}, Config: PipelineConfig{StopOnFail: true},
	})

	list1, _ := ListPipelines(p1.ID)
	if len(list1) != 2 {
		t.Errorf("project 1 expected 2 pipelines, got %d", len(list1))
	}
	list2, _ := ListPipelines(p2.ID)
	if len(list2) != 1 || list2[0].Name != "other" {
		t.Errorf("project 2 unexpected: %+v", list2)
	}
}

func TestListAllPipelines(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	p1, _ := CreateProject(CreateProjectInput{Name: "p1", Path: "/1", Language: "go"})
	p2, _ := CreateProject(CreateProjectInput{Name: "p2", Path: "/2", Language: "go"})
	_, _ = CreatePipeline(CreatePipelineInput{
		ProjectID: p1.ID, Name: "x",
		Steps: []PipelineStep{{Type: "build", Name: "s"}}, Config: PipelineConfig{StopOnFail: true},
	})
	_, _ = CreatePipeline(CreatePipelineInput{
		ProjectID: p2.ID, Name: "y",
		Steps: []PipelineStep{{Type: "build", Name: "s"}}, Config: PipelineConfig{StopOnFail: true},
	})

	all, err := ListAllPipelines()
	if err != nil {
		t.Fatalf("ListAllPipelines: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 pipelines across projects, got %d", len(all))
	}
}

// TestPipeline_CascadeDeleteFromProject 验证 FK CASCADE：删 project 时 pipelines 自动删。
// 0001 里新建的 pipelines 表带 ON DELETE CASCADE；此前老库升级的 pipelines 可能没 CASCADE，
// 本测试跑在新库上应通过。
func TestPipeline_CascadeDeleteFromProject(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	_, _ = CreatePipeline(CreatePipelineInput{
		ProjectID: pid, Name: "will-be-cascaded",
		Steps: []PipelineStep{{Type: "build", Name: "s"}}, Config: PipelineConfig{StopOnFail: true},
	})

	if err := DeleteProject(pid); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	list, _ := ListPipelines(pid)
	if len(list) != 0 {
		t.Errorf("expected pipelines cascaded on project delete, got %d", len(list))
	}
}

// TestCreatePipeline_PersistsJSONFields 验证 steps.Config 这种 map[string]interface{}
// 能正确往返（JSON marshal → DB → unmarshal）。
func TestCreatePipeline_PersistsJSONFields(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	original := map[string]interface{}{
		"tag":        "v1.2.3",
		"noCache":    true,
		"retryCount": float64(3), // JSON 数字都回来会是 float64
	}
	p, _ := CreatePipeline(CreatePipelineInput{
		ProjectID: pid, Name: "cfg",
		Steps:  []PipelineStep{{Type: "build", Name: "s", Config: original}},
		Config: PipelineConfig{StopOnFail: true},
	})

	got, _ := GetPipeline(p.ID)
	if got.Steps[0].Config["tag"] != "v1.2.3" {
		t.Errorf("Config.tag round-trip: got %v", got.Steps[0].Config["tag"])
	}
	if got.Steps[0].Config["noCache"] != true {
		t.Errorf("Config.noCache round-trip: got %v", got.Steps[0].Config["noCache"])
	}
}
