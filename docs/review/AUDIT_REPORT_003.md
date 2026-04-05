# 项目审核报告 #003

> **审查范围**: Phase 1~3 全量审查 (commits: 1876501 → 9cfbe6b)  
> **审查日期**: 2026-04-05  
> **模型**: Claude Sonnet 4.6 (Phase R 自主审查)

---

## 一、审查总览

| 维度 | 评分 | 说明 |
|------|------|------|
| 架构一致性 | ⭐⭐⭐⭐☆ 4/5 | 整体符合 local-first 原则，E2E 接入整洁 |
| 功能完整性 | ⭐⭐⭐⭐⭐ 5/5 | 所有计划任务全部完成 (23/23) |
| 代码质量 | ⭐⭐⭐⭐☆ 4/5 | 接口设计良好，发现 1 个字段名 Bug 已修复 |
| 测试覆盖 | ⭐⭐⭐⭐☆ 4/5 | 单元测试齐全，E2E 集成测试已加入 |
| 文档质量 | ⭐⭐⭐⭐⭐ 5/5 | 三份新文档结构完整 |
| 安全性 | ⭐⭐⭐⭐⭐ 5/5 | Phase 2 安全审查已修复全部已知问题 |

---

## 二、架构一致性检查

### ✅ 符合架构原则

- **Local-first**: Seahorse SQLite 是主存储，云端同步纯异步可选
- **Zero-overhead 默认禁用**: CloudMemory、TeamTool、memory_stats/health 均默认关闭
- **接口隔离**: CloudMemoryStore 接口干净，NoopStore 零开销实现 compile-time 检查
- **Config 系统集成**: 所有新功能通过 `IsToolEnabled()` + `CloudMemoryConfig` 控制
- **EventBus 订阅**: cloud_memory.go 通过事件驱动同步而非轮询

### ⚠️ 轻微偏离

- `BackupManager` 与 `SyncManager` 功能部分重叠 (两者都做 `SyncFromLocal`)
  - BackupManager = 定时全量快照
  - SyncManager = 实时事件驱动增量
  - 建议：文档中明确区分职责，避免用户困惑
- `stats.go` 中 `GetHealth()` 执行 FTS5 integrity-check 会写入 SQLite，不是纯只读

---

## 三、功能完整性检查

### ✅ 已完成 (Phase 1~3 全部)

| 功能 | 状态 | 文件 |
|------|------|------|
| TeamTool Coordinator-Worker | ✅ | `pkg/tools/team.go` |
| Termux 环境检测 | ✅ | `pkg/termux.go` |
| CloudMemoryStore 接口 | ✅ | `pkg/memory/cloud/interface.go` |
| NoopStore 零开销 | ✅ | `pkg/memory/cloud/noop.go` |
| Supabase+pgvector 后端 | ✅ | `pkg/memory/cloud/supabase.go` |
| SyncManager 异步批量同步 | ✅ | `pkg/memory/cloud/sync.go` |
| BackupManager 定时备份 | ✅ | `pkg/memory/cloud/backup.go` |
| Engine Stats/Health API | ✅ | `pkg/seahorse/stats.go` |
| memory_stats 工具 | ✅ | `pkg/seahorse/tool_stats.go` |
| memory_health 工具 | ✅ | `pkg/seahorse/tool_health.go` |
| 云记忆体 E2E 接入 | ✅ | `pkg/agent/cloud_memory.go` |
| 集成测试套件 | ✅ | `pkg/memory/cloud/integration_test.go` |

### ❌ INBOX P0 新需求（未完成，Phase 4 处理）

- 多云向量嵌入模型支持（OpenAI / Google / 阿里云 / DeepSeek / Ollama）

---

## 四、代码质量检查

### ✅ 优点

1. **TeamTool**: `sync.Once`、atomic token budget、index-based result slice 避免竞态
2. **SyncManager**: 双 `sync.Once` 防止重复 Start/Stop，优雅关闭带 10s deadline
3. **Supabase 后端**: `url.Values` 安全编码、table name regex 验证、drainAndClose 连接复用
4. **接口设计**: CloudMemoryStore 7 个方法职责清晰，可测试性强

### 🐛 发现的 Bug（已修复）

1. **P0 字段名错误**: `Memory.SessionID` 不存在，应为 `Memory.SessionKey`
   - 受影响: `pkg/agent/cloud_memory.go`, `pkg/memory/cloud/integration_test.go`
   - 状态: ✅ 已在 Phase R 中修复

### ⚠️ 技术改进建议

2. **P1**: `stats.go GetHealth()` FTS5 integrity-check 会写入 `summaries_fts/messages_fts`
   - 正确方式: `SELECT * FROM summaries_fts('integrity-check')` (只读)
   - 当前实现为 INSERT，是副作用操作

3. **P1**: `supabase.go` 的 `SimilaritySearch` 发送明文 query 文本到 Supabase RPC
   - 未来添加 embedding 时，这里需要改成先向量化再搜索
   - 接口已预留 `Memory.Embedding []float32`，好的前瞻设计

4. **P2**: `BackupManager.loop()` 初始延迟写死为 30 秒，应可配置

5. **P2**: `tool_health.go` 结果的 `issues` 如果为空，返回 `[]string{}` 而不是 `nil`
   - 已在重构时修复，JSON 序列化输出 `[]` 而非 `null`

---

## 五、技术问题清单

| 级别 | 问题 | 文件 | 状态 |
|------|------|------|------|
| P0 | Memory.SessionID 字段不存在 | cloud_memory.go, integration_test.go | ✅ 已修复 |
| P1 | GetHealth FTS5 为写操作 | seahorse/stats.go | 📋 Phase 4 任务 |
| P1 | SimilaritySearch 无向量化 | memory/cloud/supabase.go | 📋 Phase 4 主任务 |
| P2 | BackupManager 初始延迟硬编码 | memory/cloud/backup.go | 📋 可选改进 |

---

## 六、参考项目借鉴审查

### 已借鉴

| 参考来源 | 借鉴内容 | 适配 |
|----------|----------|------|
| INBOX Phase 4 指令 | 多云嵌入列表 (OpenAI/Google/阿里/DeepSeek/Ollama) | 纳入 Phase 4 |
| 现有 seahorse Store 接口 | CloudMemoryStore 设计参考 | ✅ 一致 |

### 可借鉴但未利用

| 来源 | 内容 | 建议 |
|------|------|------|
| INBOX `han-skills/memory-manager` | 多模型嵌入配置模式 | Phase 4 设计参考 |
| `pkg/providers/` 已有 LLM 接口 | 可复用 HTTP 客户端模式 | 嵌入 provider 可参考 |

---

## 七、改进建议

### P0（Phase 4 必须完成）
- **多云向量嵌入支持**: 实现 `EmbeddingProvider` 接口，支持 OpenAI / Google / 阿里云 / DeepSeek / Ollama

### P1（Phase 4 重要任务）
- **修复 GetHealth FTS5 副作用**: 改为只读 integrity check
- **EmbeddingConfig 配置系统**: 在 `CloudMemoryConfig` 中添加 embedding 配置
- **Supabase 向量化搜索**: `SimilaritySearch` 集成 embedding 后改为向量搜索

### P2（可选改进）
- **BackupManager 初始延迟可配置**: 添加 `InitialDelay` 字段
- **SyncManager 指标监控**: 添加 dropped/flushed 计数器
- **TeamTool E2E 测试**: 当前测试用 mock spawner，缺少真实 SubTurn 集成

---

## 八、Phase 4 建议

**核心主题**: 多云向量嵌入 + 记忆检索质量提升

**关键功能**:
1. `EmbeddingProvider` 接口层 + 多云实现 (OpenAI, Google, 阿里云, DeepSeek, Ollama)
2. 配置化嵌入模型切换，向后兼容
3. Supabase 后端集成向量化搜索（真正的 pgvector 语义搜索）
4. 修复 FTS5 健康检查副作用
5. 嵌入缓存层（避免重复 API 调用）

**估计任务数**: 8~10 个任务
