package pipeline

import (
	"fmt"
	"sync"
	"time"

	"github.com/flowci/flowci/internal/builder"
	"github.com/flowci/flowci/internal/deployer"
	"github.com/google/uuid"
)

type Manager struct {
	pipelines   map[string]*Pipeline
	executions  map[string]*Execution
	mu          sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		pipelines:  make(map[string]*Pipeline),
		executions: make(map[string]*Execution),
	}
}

type Pipeline struct {
	ID        string          `yaml:"id"`
	Name      string          `yaml:"name"`
	ProjectID string          `yaml:"project_id"`
	Steps     []*Step         `yaml:"steps"`
	Trigger   *TriggerConfig `yaml:"trigger"`
	Status    Status         `yaml:"status"`
	CreatedAt time.Time      `yaml:"created_at"`
	UpdatedAt time.Time      `yaml:"updated_at"`
}

type Step struct {
	ID           string            `yaml:"id"`
	Name         string            `yaml:"name"`
	Type         StepType         `yaml:"step_type"`
	Config       *StepConfig       `yaml:"config"`
	DependsOn    []string         `yaml:"depends_on"`
	Status       Status           `yaml:"status"`
	RetryCount   int              `yaml:"retry_count"`
	MaxRetries   int              `yaml:"max_retries"`
	TimeoutSec   int              `yaml:"timeout_seconds"`
}

type StepType string

const (
	StepTypeBuild       StepType = "build"
	StepTypePush        StepType = "push"
	StepTypeDeploy      StepType = "deploy"
	StepTypeScript      StepType = "script"
	StepTypeNotification StepType = "notification"
)

type StepConfig struct {
	BuildConfig     *builder.BuildConfig     `yaml:"build_config,omitempty"`
	PushConfig      *PushConfig              `yaml:"push_config,omitempty"`
	DeployConfig    *deployer.DeployConfig   `yaml:"deploy_config,omitempty"`
	ScriptContent   string                   `yaml:"script_content,omitempty"`
}

type PushConfig struct {
	ImageTag  string `yaml:"image_tag"`
	Registry  string `yaml:"registry"`
}

type TriggerConfig struct {
	Type       TriggerType `yaml:"trigger_type"`
	CronExpr  string      `yaml:"cron_expression,omitempty"`
	GitEvents []GitEvent  `yaml:"git_events,omitempty"`
}

type TriggerType string

const (
	TriggerTypeManual     TriggerType = "manual"
	TriggerTypeGitHook    TriggerType = "git_hook"
	TriggerTypeScheduled  TriggerType = "scheduled"
)

type GitEvent string

const (
	GitEventPush       GitEvent = "push"
	GitEventPullRequest GitEvent = "pull_request"
	GitEventTag        GitEvent = "tag"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning  Status = "running"
	StatusSuccess  Status = "success"
	StatusFailed   Status = "failed"
	StatusSkipped  Status = "skipped"
)

type Execution struct {
	ID          string           `yaml:"id"`
	PipelineID string           `yaml:"pipeline_id"`
	Status     Status           `yaml:"status"`
	StepResult []*StepExecution `yaml:"step_results"`
	StartedAt  time.Time       `yaml:"started_at"`
	FinishedAt *time.Time      `yaml:"finished_at,omitempty"`
	Trigger    string          `yaml:"trigger_reason"`
}

type StepExecution struct {
	StepID     string     `yaml:"step_id"`
	Status     Status     `yaml:"status"`
	StartedAt  time.Time `yaml:"started_at"`
	FinishedAt *time.Time `yaml:"finished_at,omitempty"`
	Logs       []string  `yaml:"logs"`
	Error      string     `yaml:"error,omitempty"`
}

func (m *Manager) CreatePipeline(name, projectID string) *Pipeline {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := &Pipeline{
		ID:        uuid.New().String(),
		Name:      name,
		ProjectID: projectID,
		Steps:     []*Step{},
		Trigger: &TriggerConfig{
			Type: TriggerTypeManual,
		},
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.pipelines[p.ID] = p
	return p
}

func (m *Manager) AddStep(pipelineID string, step *Step) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.pipelines[pipelineID]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	step.ID = uuid.New().String()
	p.Steps = append(p.Steps, step)
	p.UpdatedAt = time.Now()

	return nil
}

func (m *Manager) Validate(pipelineID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.pipelines[pipelineID]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	if len(p.Steps) == 0 {
		return fmt.Errorf("pipeline must have at least one step")
	}

	stepIDs := make(map[string]bool)
	for _, s := range p.Steps {
		stepIDs[s.ID] = true
	}

	for _, s := range p.Steps {
		for _, dep := range s.DependsOn {
			if !stepIDs[dep] {
				return fmt.Errorf("step %s depends on non-existent step %s", s.ID, dep)
			}
		}
	}

	if circular := m.detectCircular(p); circular != nil {
		return fmt.Errorf("circular dependency detected: %v", circular)
	}

	return nil
}

func (m *Manager) GetExecutionOrder(pipelineID string) ([][]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.pipelines[pipelineID]
	if !ok {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	inDegree := make(map[string]int)
	for _, s := range p.Steps {
		inDegree[s.ID] = 0
	}

	for _, s := range p.Steps {
		for _, dep := range s.DependsOn {
			inDegree[dep]++
		}
	}

	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var result [][]string
	for len(queue) > 0 {
		var nextLevel []string
		var nextQueue []string

		for _, stepID := range queue {
			if step, ok := p.stepMap()[stepID]; ok {
				for _, dep := range step.DependsOn {
					inDegree[dep]--
					if inDegree[dep] == 0 {
						nextQueue = append(nextQueue, dep)
					}
				}
				nextLevel = append(nextLevel, stepID)
			}
		}

		if len(nextLevel) > 0 {
			result = append(result, nextLevel)
		}
		queue = nextQueue
	}

	return result, nil
}

func (m *Manager) CreateExecution(pipelineID, trigger string) (*Execution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.pipelines[pipelineID]; !ok {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	exec := &Execution{
		ID:          uuid.New().String(),
		PipelineID:  pipelineID,
		Status:      StatusPending,
		StepResult:  []*StepExecution{},
		StartedAt:   time.Now(),
		Trigger:     trigger,
	}

	m.executions[exec.ID] = exec
	return exec, nil
}

func (m *Manager) GetExecution(execID string) (*Execution, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exec, ok := m.executions[execID]
	return exec, ok
}

func (m *Manager) GetPipeline(pipelineID string) (*Pipeline, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.pipelines[pipelineID]
	return p, ok
}

func (m *Manager) ListPipelines() []*Pipeline {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*Pipeline
	for _, p := range m.pipelines {
		list = append(list, p)
	}
	return list
}

func (p *Pipeline) stepMap() map[string]*Step {
	result := make(map[string]*Step)
	for _, s := range p.Steps {
		result[s.ID] = s
	}
	return result
}

func (m *Manager) detectCircular(p *Pipeline) []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(stepID string) []string
	dfs = func(stepID string) []string {
		visited[stepID] = true
		recStack[stepID] = true
		path = append(path, stepID)

		step := p.stepMap()[stepID]
		if step != nil {
			for _, dep := range step.DependsOn {
				if !visited[dep] {
					if result := dfs(dep); result != nil {
						return result
					}
				} else if recStack[dep] {
					path = append(path, dep)
					return path
				}
			}
		}

		path = path[:len(path)-1]
		recStack[stepID] = false
		return nil
	}

	for _, step := range p.Steps {
		if !visited[step.ID] {
			if result := dfs(step.ID); result != nil {
				return result
			}
		}
	}

	return nil
}
