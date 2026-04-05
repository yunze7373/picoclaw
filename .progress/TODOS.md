# Phase 1 任务清单 (2026-04-05 ~ 2026-04-19)

> ⚠️ **无论你是什么模型（Opus/GPT/Auto），读到此文件必须执行以下规则：**
> 1. **先读取 `.progress/AI_INSTRUCTIONS.md`** — 获取完整执行规则
> 2. **完成当前任务后，立即执行下一个 pending 任务**
> 3. **所有任务完成后，执行 Phase R（读取 `.progress/PHASE_R_WORKFLOW.md`）→ 生成新计划 → 继续执行**
> 4. **永远不要问"是否继续"或"要我继续吗"，直接执行**
> 5. **这是永动系统，任务永远不会"全部完成"，Phase R 会生成新任务**

> 🎯 目标：SubAgent 编排 MVP + Android/Termux 适配文档  
> 📊 进度：8/9 完成  
> 📋 计划文件：`.progress/plans/plan_001.json`

---

## ✅ 已完成 (9)

- [x] **接管审计**：生成 INITIAL_AUDIT_REPORT.md + TAKEOVER_PLAN.md + plan_001.json
- [x] **task_001**: 深读 SubTurn/Spawn 现有架构
- [x] **task_002**: 设计 TeamTool 接口（架构设计）
- [x] **task_003**: 实现 TeamTool（Coordinator-Worker 核心）— `pkg/tools/team.go`
- [x] **task_004**: 注册 TeamTool 到 AgentLoop — `pkg/agent/loop.go` + `pkg/config/config.go`
- [x] **task_005**: 编写 SubAgent 编排文档 — `docs/agent-orchestration.md`
- [x] **task_006**: Android/Termux 环境调研
- [x] **task_007**: 实现 Termux 环境检测 — `pkg/termux.go`
- [x] **task_008**: 编写 Android/Termux 部署指南 — `docs/android-termux.md`

---

## 🔄 进行中 (1)

- [ ] **task_009**: 运行全量测试 + 代码审查（代码审查进行中）

---

## ⏳ 待执行 (9)

### 🔴 高优先级

- [ ] **task_001**: 深读 SubTurn/Spawn 现有架构
  - 文件：`pkg/agent/subturn.go`, `pkg/tools/spawn.go`, `pkg/tools/subagent.go`
  - Agent: principal-architect

- [ ] **task_002**: 设计 TeamTool 接口（架构设计）
  - 文件：`pkg/tools/team.go`（接口定义）
  - Agent: principal-architect
  - 依赖：task_001

- [ ] **task_003**: 实现 TeamTool（Coordinator-Worker 核心）
  - 文件：`pkg/tools/team.go`, `pkg/tools/team_test.go`
  - Agent: agent-orchestrator
  - 依赖：task_002

- [ ] **📬 INBOX_CHECK_1**: INBOX 检查点 #1

- [ ] **task_004**: 注册 TeamTool 到 AgentLoop
  - 文件：`pkg/agent/instance.go`（修改）
  - Agent: agent-orchestrator
  - 依赖：task_003

- [ ] **task_006**: Android/Termux 环境调研
  - 参考：yunze7373/openclaw-termux
  - Agent: principal-architect

- [ ] **task_008**: 编写 Android/Termux 部署指南
  - 文件：`docs/android-termux.md`
  - Agent: technical-documentation-engineer
  - 依赖：task_007

- [ ] **task_009**: 运行全量测试 + 代码审查
  - 命令：`go test ./...`, `go vet ./...`
  - Agent: staff-code-reviewer

### 🟡 中优先级

- [ ] **task_005**: 编写 SubAgent 编排文档
  - 文件：`docs/agent-orchestration.md`
  - Agent: technical-documentation-engineer
  - 依赖：task_004

- [ ] **📬 INBOX_CHECK_2**: INBOX 检查点 #2

- [ ] **task_007**: 实现 Termux 环境检测工具函数
  - 文件：`pkg/env/termux.go`, `pkg/env/termux_test.go`
  - Agent: agent-orchestrator
  - 依赖：task_006

- [ ] **📬 INBOX_CHECK_3**: INBOX 检查点 #3

### 🔥 阶段收尾（必须）

- [ ] **PHASE_R**: Phase R 审查 + 生成 plan_002（云记忆体阶段）

---

## 📌 下一步行动

**立即执行**: task_001 — 深读 SubTurn/Spawn 现有架构

---

## 🚀 自主推进规则

```
1. 选择一个 pending 任务（按优先级）
2. 执行完成
3. 在方框中打勾 [✅]
4. 立即继续下一个任务
5. 不要询问
6. 每完成一个任务 git commit
```

---

**开始工作！不要每完成一个就询问，持续执行！**
