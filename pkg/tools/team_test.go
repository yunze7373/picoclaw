package tools

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// teamMockSpawner returns a canned response after a configurable delay.
type teamMockSpawner struct {
	delay    time.Duration
	failAt   int // index that should fail (-1 = none)
	callCount atomic.Int32
}

func (m *teamMockSpawner) SpawnSubTurn(ctx context.Context, cfg SubTurnConfig) (*ToolResult, error) {
	idx := int(m.callCount.Add(1)) - 1

	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.failAt >= 0 && idx == m.failAt {
		return nil, errors.New("simulated worker failure")
	}

	// Deduct from token budget if set
	if cfg.InitialTokenBudget != nil {
		cfg.InitialTokenBudget.Add(-100)
	}

	task := cfg.SystemPrompt
	if parts := strings.SplitN(task, "Task: ", 2); len(parts) == 2 {
		task = parts[1]
	}
	return &ToolResult{
		ForLLM: "Worker result: " + task,
	}, nil
}

func TestTeamTool_Name(t *testing.T) {
	tool := NewTeamTool(nil, "test-model", 4096, 0.7)
	if tool.Name() != "team_create" {
		t.Errorf("Expected name 'team_create', got %q", tool.Name())
	}
}

func TestTeamTool_Execute_NoSpawner(t *testing.T) {
	tool := NewTeamTool(nil, "test-model", 4096, 0.7)
	result := tool.Execute(context.Background(), map[string]any{
		"workers": []any{
			map[string]any{"task": "hello"},
		},
	})
	if !result.IsError {
		t.Fatal("Expected error when spawner is nil")
	}
	if !strings.Contains(result.ForLLM, "spawner not configured") {
		t.Errorf("Unexpected error message: %s", result.ForLLM)
	}
}

func TestTeamTool_Execute_EmptyWorkers(t *testing.T) {
	spawner := &teamMockSpawner{}
	tool := NewTeamTool(spawner, "test-model", 4096, 0.7)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"missing workers", map[string]any{}},
		{"empty array", map[string]any{"workers": []any{}}},
		{"wrong type", map[string]any{"workers": "not-an-array"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tool.Execute(context.Background(), tt.args)
			if !result.IsError {
				t.Fatal("Expected error for invalid workers")
			}
		})
	}
}

func TestTeamTool_Execute_WorkerMissingTask(t *testing.T) {
	spawner := &teamMockSpawner{}
	tool := NewTeamTool(spawner, "test-model", 4096, 0.7)

	result := tool.Execute(context.Background(), map[string]any{
		"workers": []any{
			map[string]any{"label": "no-task"},
		},
	})
	if !result.IsError {
		t.Fatal("Expected error for worker without task")
	}
	if !strings.Contains(result.ForLLM, "task is required") {
		t.Errorf("Unexpected error message: %s", result.ForLLM)
	}
}

func TestTeamTool_Execute_SingleWorker(t *testing.T) {
	spawner := &teamMockSpawner{failAt: -1}
	tool := NewTeamTool(spawner, "test-model", 4096, 0.7)

	result := tool.Execute(context.Background(), map[string]any{
		"workers": []any{
			map[string]any{"task": "analyse code", "label": "researcher"},
		},
	})
	if result.IsError {
		t.Fatalf("Unexpected error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "1 worker(s)") {
		t.Errorf("Expected 1 worker in output, got: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "[researcher]") {
		t.Errorf("Expected label in output, got: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "1/1 workers succeeded") {
		t.Errorf("Expected success summary, got: %s", result.ForLLM)
	}
}

func TestTeamTool_Execute_MultipleWorkers(t *testing.T) {
	spawner := &teamMockSpawner{failAt: -1}
	tool := NewTeamTool(spawner, "test-model", 4096, 0.7)

	result := tool.Execute(context.Background(), map[string]any{
		"workers": []any{
			map[string]any{"task": "research APIs", "label": "researcher", "phase": "research"},
			map[string]any{"task": "write code", "label": "coder", "phase": "implementation"},
			map[string]any{"task": "verify output", "label": "verifier", "phase": "verification"},
		},
	})
	if result.IsError {
		t.Fatalf("Unexpected error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "3 worker(s)") {
		t.Errorf("Expected 3 workers: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "3/3 workers succeeded") {
		t.Errorf("Expected all succeeded: %s", result.ForLLM)
	}
	for _, label := range []string{"[researcher]", "[coder]", "[verifier]"} {
		if !strings.Contains(result.ForLLM, label) {
			t.Errorf("Expected label %s in output", label)
		}
	}
	for _, phase := range []string{"(phase: research)", "(phase: implementation)", "(phase: verification)"} {
		if !strings.Contains(result.ForLLM, phase) {
			t.Errorf("Expected phase %s in output", phase)
		}
	}
}

func TestTeamTool_Execute_PartialFailure(t *testing.T) {
	spawner := &teamMockSpawner{failAt: 1} // second worker fails
	tool := NewTeamTool(spawner, "test-model", 4096, 0.7)

	result := tool.Execute(context.Background(), map[string]any{
		"workers": []any{
			map[string]any{"task": "task A"},
			map[string]any{"task": "task B"},
			map[string]any{"task": "task C"},
		},
	})
	// Partial failure is not a tool error — we report individual results
	if result.IsError {
		t.Fatalf("Unexpected tool-level error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "ERROR:") {
		t.Errorf("Expected ERROR marker for failed worker: %s", result.ForLLM)
	}
	// At least 2 should succeed (workers 0 and 2)
	if !strings.Contains(result.ForLLM, "2/3 workers succeeded") {
		t.Errorf("Expected 2/3 succeeded: %s", result.ForLLM)
	}
}

func TestTeamTool_Execute_SharedTokenBudget(t *testing.T) {
	spawner := &teamMockSpawner{failAt: -1}
	tool := NewTeamTool(spawner, "test-model", 4096, 0.7)

	result := tool.Execute(context.Background(), map[string]any{
		"workers": []any{
			map[string]any{"task": "task 1"},
			map[string]any{"task": "task 2"},
		},
		"token_budget": float64(1000),
	})
	if result.IsError {
		t.Fatalf("Unexpected error: %s", result.ForLLM)
	}
	// Each mock worker deducts 100 tokens, so 1000 - 200 = 800
	if !strings.Contains(result.ForLLM, "Token budget remaining: 800") {
		t.Errorf("Expected token budget remaining 800, got: %s", result.ForLLM)
	}
}

func TestTeamTool_Execute_CustomTimeout(t *testing.T) {
	spawner := &teamMockSpawner{delay: 50 * time.Millisecond, failAt: -1}
	tool := NewTeamTool(spawner, "test-model", 4096, 0.7)

	result := tool.Execute(context.Background(), map[string]any{
		"workers": []any{
			map[string]any{"task": "slow task"},
		},
		"timeout_seconds": float64(10),
	})
	if result.IsError {
		t.Fatalf("Unexpected error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "1/1 workers succeeded") {
		t.Errorf("Expected success: %s", result.ForLLM)
	}
}

func TestBuildWorkerPrompt_WithPhase(t *testing.T) {
	prompt := buildWorkerPrompt("analyse code", "researcher", "research")

	if !strings.Contains(prompt, "<task_notification>") {
		t.Error("Expected XML task_notification tag")
	}
	if !strings.Contains(prompt, "<phase>research</phase>") {
		t.Error("Expected phase tag")
	}
	if !strings.Contains(prompt, "<label>researcher</label>") {
		t.Error("Expected label tag")
	}
	if !strings.Contains(prompt, "Task: analyse code") {
		t.Error("Expected task text")
	}
}

func TestBuildWorkerPrompt_NoPhase(t *testing.T) {
	prompt := buildWorkerPrompt("do something", "worker1", "")

	if strings.Contains(prompt, "<task_notification>") {
		t.Error("Should not have XML tags when phase is empty")
	}
	if !strings.Contains(prompt, "Task: do something") {
		t.Error("Expected task text")
	}
}
