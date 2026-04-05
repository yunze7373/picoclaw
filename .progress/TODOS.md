# 📋 TODOS - Phase 2: Cloud Memory & Analysis Tools

> ⚠️ **无论你是什么模型（Opus/GPT/Auto），读到此文件必须执行以下规则：**
> 1. **先读取 `.progress/AI_INSTRUCTIONS.md`** — 获取完整执行规则
> 2. **完成当前任务后，立即执行下一个 pending 任务**
> 3. **所有任务完成后，执行 Phase R → 生成新计划 → 继续执行**
> 4. **永远不要问"是否继续"或"要我继续吗"，直接执行**
> 5. **这是永动系统，任务永远不会"全部完成"，Phase R 会生成新任务**

> 🎯 目标：可选云记忆体模块 + 记忆分析工具  
> 📊 进度：8/8 完成  
> 📋 计划文件：`.progress/plans/plan_002.json`

---

## ✅ Phase 1 已完成 (9/9)

- [x] task_001~009: SubAgent 编排 MVP + Android/Termux 适配（详见 AUDIT_REPORT_002）

---

## 🔄 Phase 2 任务 (当前)

### 🔴 高优先级（可并行启动）

- [x] **task_010**: 设计 CloudMemoryStore 接口
  - 文件：`pkg/memory/cloud/interface.go`, `pkg/memory/cloud/noop.go`
  - 基于现有 `pkg/memory/store.go` Store 接口扩展

- [x] **task_014**: 实现 memory_stats 工具
  - 文件：`pkg/seahorse/tool_stats.go`
  - 查询 seahorse SQLite DB 获取统计信息

### 🟡 中优先级（依赖 task_010）

- [x] **task_011**: 实现 Supabase pgvector 后端 [依赖: 010]
  - 文件：`pkg/memory/cloud/supabase.go`
  - 使用 net/http，无外部依赖

- [x] **task_012**: 添加云记忆体配置 [依赖: 010]
  - 文件：`pkg/config/config.go` 扩展

- [x] **task_015**: 实现 memory_health 工具 [依赖: 014]
  - 文件：`pkg/seahorse/tool_health.go`

### 🟢 后续任务

- [x] **task_013**: EventBus 同步集成 [依赖: 011, 012]
  - 文件：`pkg/memory/cloud/sync.go`

- [x] **task_016**: 编写云记忆体文档 [依赖: 013]
  - 文件：`docs/cloud-memory.md`

- [x] **task_017**: 测试 + 代码审查 [依赖: 013, 015, 016] — ✅ 完成

### 📬 检查点

- [ ] **inbox_check_002**: INBOX 检查点 [在 task_014 后]
- [ ] **phase_r_trigger_002**: Phase R 触发 [在 task_017 后]

---

## 📌 下一步行动

**立即并行执行**: task_010 + task_014（无依赖关系）

---

**开始工作！持续执行！**
