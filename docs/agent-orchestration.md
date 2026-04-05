# Agent Orchestration: Team Tool (Coordinator-Worker Pattern)

> PicoClaw v0.3+ | Feature Flag: `tools.team_create.enabled = true`

PicoClaw supports multi-agent orchestration through the **`team_create`** tool, which implements a **Coordinator-Worker** pattern on top of the existing SubTurn system.

## Quick Start

Enable the tool in your config:

```toml
[tools.team_create]
enabled = true

# Also requires subagent support:
[tools.subagent]
enabled = true
[tools.spawn]
enabled = true
```

## How It Works

```
┌────────────────────────────────┐
│  Coordinator Agent             │
│  (calls team_create tool)      │
│                                │
│  ┌──────┐ ┌──────┐ ┌──────┐   │
│  │Worker│ │Worker│ │Worker│   │  ← Fan-out (parallel)
│  │  1   │ │  2   │ │  3   │   │
│  └──┬───┘ └──┬───┘ └──┬───┘   │
│     │        │        │        │
│     └────────┼────────┘        │
│              ▼                 │
│     Merged Results             │  ← Fan-in (wait all)
└────────────────────────────────┘
```

The coordinator calls `team_create` with a list of worker specifications. Each worker is spawned as an independent SubTurn running concurrently. The tool blocks until all workers complete, then returns the merged results.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `workers` | array | ✅ | List of worker task objects |
| `workers[].task` | string | ✅ | Task description for the worker |
| `workers[].label` | string | ❌ | Human-readable label (e.g., "researcher") |
| `workers[].phase` | string | ❌ | Workflow phase: `research`, `synthesis`, `implementation`, `verification` |
| `token_budget` | integer | ❌ | Shared token budget across all workers |
| `timeout_seconds` | integer | ❌ | Per-worker timeout (default: 300s) |

## Example: Four-Phase Workflow

```json
{
  "workers": [
    {
      "task": "Research existing API designs for pagination. Find at least 3 different approaches.",
      "label": "researcher",
      "phase": "research"
    },
    {
      "task": "Based on research findings, synthesize a recommendation for our pagination API.",
      "label": "synthesizer",
      "phase": "synthesis"
    },
    {
      "task": "Implement the cursor-based pagination in pkg/api/pagination.go",
      "label": "implementer",
      "phase": "implementation"
    },
    {
      "task": "Verify the pagination implementation handles edge cases: empty results, single page, last page.",
      "label": "verifier",
      "phase": "verification"
    }
  ],
  "token_budget": 50000,
  "timeout_seconds": 600
}
```

> **Note:** In this example all four phases run in parallel. For sequential workflows where later phases depend on earlier results, call `team_create` multiple times — once per stage.

## XML Task Notification Format

When a `phase` is specified, workers receive their task wrapped in a structured XML format:

```xml
<task_notification>
  <phase>research</phase>
  <label>researcher</label>
  <payload>Research existing API designs for pagination...</payload>
</task_notification>
```

This enables workers to understand their role in the overall workflow and adapt their behavior accordingly.

## Token Budget

The shared token budget is an `atomic.Int64` counter passed to all workers via `InitialTokenBudget`. Each worker's LLM calls deduct from this shared pool. When the budget reaches zero, workers will stop making LLM calls.

```json
{
  "workers": [
    {"task": "quick analysis", "label": "fast-worker"},
    {"task": "deep dive",      "label": "deep-worker"}
  ],
  "token_budget": 10000
}
```

## Difference from `spawn` Tool

| Feature | `spawn` | `team_create` |
|---------|---------|---------------|
| Execution | Fire-and-forget (async) | Blocks until all complete |
| Workers | Single | Multiple parallel |
| Token budget | Per-worker | Shared across team |
| Result delivery | Via `spawn_status` polling | Inline merged results |
| Use case | Background tasks | Coordinated parallel work |

## Resource Considerations

- Each worker is a SubTurn with an ephemeral (in-memory) session
- Workers inherit the parent's tool registry (minus spawn tools to prevent recursion)
- Default concurrency limit: 5 workers (configurable via `agents.defaults.subturn.max_concurrent`)
- Default depth limit: 3 levels (team workers count as depth +1)
- On resource-constrained devices (Android/Termux), consider using smaller token budgets and fewer workers

## Configuration Reference

```toml
[tools.team_create]
enabled = true    # Enable the team_create tool (default: false)

[agents.defaults.subturn]
max_concurrent = 5           # Max parallel SubTurns (affects team workers)
max_depth = 3                # Max SubTurn nesting depth
default_timeout_minutes = 5  # Default per-worker timeout
default_token_budget = 0     # 0 = no default budget
```

---

*See also: [SubTurn Documentation](subturn.md) | [Spawn Tasks](spawn-tasks.md)*
