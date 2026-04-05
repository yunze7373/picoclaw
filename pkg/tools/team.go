package tools

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// TeamTool implements a Coordinator-Worker pattern on top of the existing
// SubTurn system.  A coordinator agent can call this tool to fan-out tasks
// to multiple parallel worker sub-agents, wait for all of them to complete,
// and receive a merged result.
//
// It reuses SubTurnSpawner (the same machinery behind the "spawn" tool) but
// adds synchronous fan-out / fan-in semantics plus a shared token budget so
// the total cost of all workers stays bounded.
type TeamTool struct {
	spawner      SubTurnSpawner
	defaultModel string
	maxTokens    int
	temperature  float64
}

// Compile-time check: TeamTool implements Tool.
var _ Tool = (*TeamTool)(nil)

// NewTeamTool creates a TeamTool backed by the given SubTurnSpawner.
func NewTeamTool(spawner SubTurnSpawner, defaultModel string, maxTokens int, temperature float64) *TeamTool {
	return &TeamTool{
		spawner:      spawner,
		defaultModel: defaultModel,
		maxTokens:    maxTokens,
		temperature:  temperature,
	}
}

// SetSpawner replaces the SubTurnSpawner.
// Must be called before any concurrent Execute calls (during setup only).
func (t *TeamTool) SetSpawner(spawner SubTurnSpawner) {
	t.spawner = spawner
}

func (t *TeamTool) Name() string { return "team_create" }

func (t *TeamTool) Description() string {
	return `Create a team of parallel worker sub-agents, each assigned a task. ` +
		`All workers run concurrently and share a token budget. ` +
		`The tool blocks until every worker finishes and returns all results. ` +
		`Use this for tasks that decompose into independent sub-problems ` +
		`(research, synthesis, implementation, verification).`
}

func (t *TeamTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"workers": map[string]any{
				"type":        "array",
				"description": "List of worker task descriptions. Each entry spawns one parallel sub-agent.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"task": map[string]any{
							"type":        "string",
							"description": "The task for this worker to complete.",
						},
						"label": map[string]any{
							"type":        "string",
							"description": "Short human-readable label for the worker (e.g. 'researcher', 'verifier').",
						},
						"phase": map[string]any{
							"type":        "string",
							"description": "Optional workflow phase: research | synthesis | implementation | verification.",
							"enum":        []string{"research", "synthesis", "implementation", "verification"},
						},
					},
					"required": []string{"task"},
				},
			},
			"token_budget": map[string]any{
				"type":        "integer",
				"description": "Optional shared token budget across all workers. Default: no limit.",
			},
			"timeout_seconds": map[string]any{
				"type":        "integer",
				"description": "Optional per-worker timeout in seconds. Default: 300 (5 minutes).",
			},
		},
		"required": []string{"workers"},
	}
}

func (t *TeamTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.spawner == nil {
		return ErrorResult("team_create: spawner not configured")
	}

	// Parse workers
	rawWorkers, ok := args["workers"]
	if !ok {
		return ErrorResult("team_create: 'workers' parameter is required")
	}
	workerList, ok := rawWorkers.([]any)
	if !ok || len(workerList) == 0 {
		return ErrorResult("team_create: 'workers' must be a non-empty array")
	}

	type workerSpec struct {
		task  string
		label string
		phase string
	}
	specs := make([]workerSpec, 0, len(workerList))
	for i, raw := range workerList {
		m, ok := raw.(map[string]any)
		if !ok {
			return ErrorResult(fmt.Sprintf("team_create: worker[%d] must be an object", i))
		}
		task, _ := m["task"].(string)
		if strings.TrimSpace(task) == "" {
			return ErrorResult(fmt.Sprintf("team_create: worker[%d].task is required", i))
		}
		label, _ := m["label"].(string)
		phase, _ := m["phase"].(string)
		specs = append(specs, workerSpec{task: task, label: label, phase: phase})
	}

	// Cap the number of workers to prevent resource exhaustion
	const maxWorkers = 20
	if len(specs) > maxWorkers {
		return ErrorResult(fmt.Sprintf("team_create: too many workers (%d), maximum is %d", len(specs), maxWorkers))
	}

	// Parse optional token budget
	var sharedBudget *atomic.Int64
	if rawBudget, ok := args["token_budget"]; ok {
		if budgetFloat, ok := rawBudget.(float64); ok && budgetFloat > 0 {
			sharedBudget = &atomic.Int64{}
			sharedBudget.Store(int64(budgetFloat))
		}
	}

	// Parse optional timeout
	timeout := 5 * time.Minute
	if rawTimeout, ok := args["timeout_seconds"]; ok {
		if ts, ok := rawTimeout.(float64); ok && ts > 0 {
			timeout = time.Duration(ts) * time.Second
		}
	}

	// Fan-out: launch all workers concurrently
	type workerResult struct {
		index int
		label string
		phase string
		res   *ToolResult
		err   error
	}

	var wg sync.WaitGroup
	results := make([]workerResult, len(specs))

	for i, spec := range specs {
		wg.Add(1)
		go func(idx int, s workerSpec) {
			defer wg.Done()

			systemPrompt := buildWorkerPrompt(s.task, s.label, s.phase)

			cfg := SubTurnConfig{
				Model:              t.defaultModel,
				SystemPrompt:       systemPrompt,
				MaxTokens:          t.maxTokens,
				Temperature:        t.temperature,
				Async:              false, // synchronous — we collect inline
				Critical:           false,
				Timeout:            timeout,
				InitialTokenBudget: sharedBudget,
			}

			res, err := t.spawner.SpawnSubTurn(ctx, cfg)
			results[idx] = workerResult{
				index: idx,
				label: s.label,
				phase: s.phase,
				res:   res,
				err:   err,
			}
		}(i, spec)
	}

	// Fan-in: wait for all workers
	wg.Wait()

	// Merge results
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Team completed: %d worker(s)\n\n", len(results)))

	successCount := 0
	for _, wr := range results {
		header := fmt.Sprintf("--- Worker %d", wr.index+1)
		if wr.label != "" {
			header += fmt.Sprintf(" [%s]", wr.label)
		}
		if wr.phase != "" {
			header += fmt.Sprintf(" (phase: %s)", wr.phase)
		}
		header += " ---\n"
		sb.WriteString(header)

		if wr.err != nil {
			sb.WriteString(fmt.Sprintf("ERROR: %v\n", wr.err))
		} else if wr.res != nil {
			sb.WriteString(wr.res.ForLLM)
			sb.WriteString("\n")
			successCount++
		} else {
			sb.WriteString("(no result)\n")
		}
		sb.WriteString("\n")
	}

	if sharedBudget != nil {
		remaining := sharedBudget.Load()
		if remaining < 0 {
			remaining = 0
		}
		sb.WriteString(fmt.Sprintf("Token budget remaining: %d\n", remaining))
	}

	sb.WriteString(fmt.Sprintf("Summary: %d/%d workers succeeded", successCount, len(results)))

	return NewToolResult(sb.String())
}

// buildWorkerPrompt generates the system prompt for a worker sub-agent,
// including the optional phase and label metadata as XML for structured parsing.
func buildWorkerPrompt(task, label, phase string) string {
	var sb strings.Builder
	sb.WriteString("You are a worker sub-agent in a team. Complete the assigned task independently and report your findings.\n\n")

	if phase != "" {
		sb.WriteString(fmt.Sprintf("<task_notification>\n  <phase>%s</phase>\n", escapeXML(phase)))
		if label != "" {
			sb.WriteString(fmt.Sprintf("  <label>%s</label>\n", escapeXML(label)))
		}
		sb.WriteString(fmt.Sprintf("  <payload>%s</payload>\n</task_notification>\n\n", escapeXML(task)))
	}

	sb.WriteString(fmt.Sprintf("Task: %s", task))
	return sb.String()
}

// escapeXML escapes XML special characters in a string.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
