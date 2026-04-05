# 🚀 PicoClaw Contributor 接管计划 (TAKEOVER_PLAN)

> 版本：v1.0 | 创建：2026-04-05  
> 基于：`.progress/reports/INITIAL_AUDIT_REPORT.md` + `.progress/PROJECT_IDEA.md`

---

## 接管目标

作为 PicoClaw 贡献者，在不破坏现有架构的前提下，增强多 Agent 协作能力、添加可选云记忆体、优化 Android/Termux 体验，并建立记忆分析工具。

**核心约束**：
- 所有新功能**默认关闭**，可选启用
- 不推翻现有 SubTurn/LCM/MCP 系统
- 保持跨平台兼容（Linux/Android/macOS/Windows + RISC-V/ARM/MIPS/x86）
- 基础 RAM 增量 < 5MB（启用可选功能上限 <50MB）

---

## 主要发力点

### 🎯 发力点 1：增强 SubAgent 编排系统（MVP，高优先级）

**现状**：`pkg/tools/spawn.go` 已有 `spawn` 工具，`pkg/agent/subturn.go` 支持同步/异步子 Agent。

**目标**：在现有基础上扩展 Coordinator-Worker 模式。

**关键设计**：
```go
// pkg/tools/team.go - 新增 team_create 工具
// 支持 coordinator 分配 worker 任务，收集结果
type TeamTool struct { ... }

// pkg/tools/message.go - Agent 间消息传递（已有基础设施）
// 利用现有 EventBus 实现跨 Agent 消息

// XML 任务通知格式（借鉴 claude-code）
// <task_notification>
//   <task_id>xxx</task_id>
//   <phase>research|synthesis|implementation|verification</phase>
//   <payload>...</payload>
// </task_notification>
```

**交付物**：
- `pkg/tools/team.go` - TeamCreate 工具
- `docs/agent-orchestration.md` - 编排使用文档
- 测试覆盖率 >80%

---

### 🎯 发力点 2：Android/Termux 适配文档与配置（MVP，高优先级）

**现状**：Android APK 已发布，但缺乏 Termux 使用指南。

**目标**：完整的 Android/Termux 配置文档 + 可选的 Termux 检测逻辑。

**关键内容**：
- Termux 路径适配 (`$PREFIX` 环境检测)
- ARM64 原生模块编译指南（针对 sqlite-vec）
- 资源约束模式配置示例
- 权限无需 root 的配置说明

**交付物**：
- `docs/android-termux.md` - 完整 Termux 部署指南
- `pkg/env/termux.go` - Termux 环境检测工具函数（可选）

---

### 🎯 发力点 3：可选云记忆体模块（Phase 2）

**现状**：仅有 SQLite + Markdown 本地记忆，无向量搜索。

**设计原则**：**零侵入**。默认关闭时完全不影响现有系统。

**接口设计**：
```go
// pkg/memory/cloud/interface.go
type CloudMemoryStore interface {
    UpsertMemory(ctx context.Context, m Memory) error
    SimilaritySearch(ctx context.Context, query string, topK int) ([]Memory, error)
    SyncToCloud(ctx context.Context) error
    HealthCheck(ctx context.Context) error
}

// 配置开关
// [tools.memory.cloud]
// enabled = false  # 默认关闭
// backend = "supabase"  # supabase | none
// connection_string = "..."
```

**技术决策**：向量嵌入使用本地轻量模型（避免强依赖外部服务）或 Supabase pgvector，由用户配置决定。

**交付物**：
- `pkg/memory/cloud/` - 云记忆体接口 + Supabase 实现
- `config/` - 云记忆体配置示例
- `docs/cloud-memory.md` - 配置文档

---

### 🎯 发力点 4：记忆分析工具（Phase 3）

**现状**：Seahorse LCM 内部有 Token 计数，但无对外暴露的统计接口。

**目标**：作为可选 Skill 或 MCP 工具暴露记忆统计。

**功能**：
- `memory_stats` 工具：返回记忆条数、Token 使用、压缩历史
- `memory_health` 工具：检查数据库完整性、WAL 状态
- Cron Job 定时备份（利用现有 `pkg/cron/` 系统）

**交付物**：
- `pkg/tools/memory_stats.go` - 记忆统计工具
- `pkg/seahorse/stats.go` - 统计接口暴露

---

### 🎯 发力点 5：Browser Control 轻量化 Skill（Phase 3，低优先级）

作为可选 Skill，不依赖完整 Chrome，支持基本截图和 DOM 查询。延后至 Phase 3 评估。

---

## 重构建议

1. **统一 SQLite 驱动策略**：明确 `modernc.org/sqlite`（纯 Go，无 CGO）为默认，`mattn/go-sqlite3` 用于需要扩展（sqlite-vec）的构建 tag
2. **RAM 基准测试**：在添加新功能前，建立 RAM 使用基准（当前 10-20MB 需要验证）
3. **SubTurn 深度限制**：当前默认 3 层深度，Coordinator-Worker 模式需要验证是否足够

---

## 里程碑

| 里程碑 | 内容 | 目标完成 |
|--------|------|---------|
| M1 MVP | SubAgent 编排增强 + Android 文档 | Phase 1 |
| M2 云记忆体 | 可选云记忆体模块（接口+Supabase 实现） | Phase 2 |
| M3 分析工具 | 记忆分析 + Browser Control | Phase 3 |

---

## 技术风险登记

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| sqlite-vec 需要 CGO | 破坏纯 Go 编译 | 使用构建 tag 隔离，默认不启用 |
| SubAgent 编排增加内存 | 超出 50MB 限制 | 严格测试，worker 使用 ephemeral session |
| Supabase 外部依赖 | 离线不可用 | 完全可选，有降级逻辑 |
| Android 路径兼容 | 运行时崩溃 | 在 CI 中添加 Termux 环境测试 |
