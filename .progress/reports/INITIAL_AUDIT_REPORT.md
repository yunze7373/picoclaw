# 📋 PicoClaw 接管初始审计报告 (INITIAL_AUDIT_REPORT)

> 生成时间：2026-04-05  
> 审计人：Copilot 接管引擎 (Takeover Engine)

---

## 1. 系统架构推断

### 整体架构
PicoClaw 是一个以 Go 语言编写的超轻量级 AI 助手框架，采用**单二进制、事件驱动、模块化插件**架构。

```
┌─────────────────────────────────────────────────────────────────┐
│  Channels Layer（输入/输出渠道）                                  │
│  Telegram / Discord / Slack / Matrix / IRC / WeChat / DingTalk   │
│  WeCom / Lark / WhatsApp / VK / QQ / Line / Web                  │
└──────────────────────────────┬──────────────────────────────────┘
                               │ EventBus
┌──────────────────────────────▼──────────────────────────────────┐
│  Agent Core (pkg/agent/)                                         │
│  ┌───────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │ AgentLoop │  │ SubTurn      │  │ Hooks / Steering / Events │  │
│  │ (loop.go) │  │ (subturn.go) │  │ (hooks.go / steering.go)  │  │
│  └───────────┘  └──────────────┘  └──────────────────────────┘  │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  Tools Layer (pkg/tools/)                                        │
│  spawn / subagent / spawn_status / shell / web / edit / MCP      │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  Memory System                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Seahorse LCM (pkg/seahorse/) - Short-Term Memory        │    │
│  │  SQLite WAL + 256-shard 分片锁 + LLM 驱动压缩            │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │  Memory Store (pkg/memory/) - Long-Term + Session        │    │
│  │  JSONL 会话历史 + Markdown 长期记忆                       │    │
│  └─────────────────────────────────────────────────────────┘    │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  Providers (pkg/providers/)                                      │
│  Anthropic / OpenAI / AWS Bedrock / Azure / Kimi / MiMo / etc.   │
└─────────────────────────────────────────────────────────────────┘
```

### 关键特性
- **单二进制**：Go 编译，零 runtime 依赖
- **资源极限**：基础 <10MB RAM（目标）
- **跨平台**：x86_64 / ARM64 / MIPS / RISC-V / LoongArch
- **Android 已支持**：APK 已发布（2026-03-31）
- **MCP 协议**：Go SDK 支持（v0.2.1+）
- **SubTurn 系统**：支持同步/异步子 Agent，深度限制 3 层，并发上限 5

---

## 2. 技术栈

| 层次 | 技术 |
|------|------|
| 语言 | Go 1.25.8 |
| 数据库 | SQLite (`modernc.org/sqlite` 纯 Go + `mattn/go-sqlite3` CGO) |
| AI SDK | Anthropic SDK, OpenAI SDK v3, AWS Bedrock, Azure, GitHub Copilot SDK |
| MCP | `modelcontextprotocol/go-sdk v1.4.1` |
| 渠道 SDK | Telegram, Discord, Slack, Matrix, IRC, WhatsApp, VK, QQ, DingTalk, Lark |
| 配置 | TOML (`BurntSushi/toml`) |
| CLI | `spf13/cobra` |
| 日志 | `rs/zerolog` |
| 测试 | `stretchr/testify` |
| 音视频 | `pion/webrtc`, `pion/rtp` |
| 系统托盘 | `fyne.io/systray` |
| 定时任务 | `adhocore/gronx` |
| TUI | `rivo/tview`, `gdamore/tcell` |

---

## 3. 历史包袱与未竟事业

### 3.1 未竟的 SubAgent 编排
- **现状**：`pkg/tools/spawn.go` 和 `pkg/tools/subagent.go` 已有基础 spawn/subagent 工具
- **缺失**：没有 Coordinator-Worker 模式，无 Research/Synthesis/Implementation/Verification 四阶段工作流
- **缺失**：没有跨 Agent 消息通道（仅靠 `pendingResults` channel），无 `TeamCreate` / `SendMessage` 等高级编排工具
- **文档**：`docs/subturn.md` 和 `docs/spawn-tasks.md` 存在但架构文档不完整

### 3.2 云记忆体完全缺失
- **现状**：Seahorse LCM（SQLite WAL）是唯一短期记忆，`pkg/memory/` 提供 JSONL + Markdown 长期记忆
- **缺失**：完全没有向量搜索、语义检索、云同步等云记忆体功能
- **缺失**：`modernc.org/sqlite` 纯 Go 实现不支持 `sqlite-vec`，向量扩展需要 CGO

### 3.3 Android/Termux 优化不完整
- **现状**：已发布 Android APK，基础支持存在
- **缺失**：Termux 路径 (`$PREFIX/home`) 适配文档；ARM64 原生模块编译指南；资源约束模式（<50MB RAM 目标未验证）
- **缺失**：无 Android 专用配置文档

### 3.4 记忆分析工具缺失
- **现状**：`pkg/seahorse/` 有 Token 计数（`short_constants.go`），有 Compaction Engine
- **缺失**：记忆健康检查 API；统计面板（总记忆数/Token 使用率）；定时备份 Cron Job

### 3.5 Browser Control 缺失
- **现状**：无 CDP/Browser 相关代码
- **缺失**：轻量级浏览器控制 Skill（截图、DOM 查询）

### 3.6 代码质量隐患
- README 提示：最近合并多个 PR 后 RAM 使用量上升至 10-20MB，"resource optimization is planned after feature stabilization"
- `go.mod` 中 `replace` 了 discordgo（使用第三方 fork）
- 存在两套 SQLite 驱动（CGO + 纯 Go），可能引起平台兼容问题

---

## 4. PROJECT_IDEA 期望 vs 当前代码差距分析

| 期望功能 | 当前状态 | 差距 |
|---------|---------|------|
| Coordinator-Worker SubAgent 编排 | spawn/subagent 基础存在，无高级编排 | 🔴 需要设计+实现 |
| Research/Synthesis/Implementation/Verification 四阶段 | 无 | 🔴 全新功能 |
| 跨 Agent XML 消息格式 | 无 | 🔴 全新功能 |
| 云记忆体 (Supabase+pgvector，可选) | 完全缺失 | 🔴 需要设计+实现 |
| 向量语义搜索 | 无（SQLite 无 sqlite-vec 扩展） | 🔴 需评估方案 |
| 跨设备记忆同步 | 无 | 🔴 需要设计 |
| Termux 路径适配 | 基础 Android 支持存在 | 🟡 需文档+测试 |
| ARM64 从源码编译 sqlite-vec | 无 | 🟡 可行但需工作 |
| 资源占用优化 <50MB | 目前 10-20MB（功能变多后） | 🟢 趋势向好 |
| 记忆健康检查 | 无 | 🟡 相对容易实现 |
| 统计面板 | 无 | 🟡 需要新接口 |
| 定时备份 Cron Job | cron 系统已存在 (`pkg/cron/`) | 🟢 基础设施已有 |
| Browser Control Skill（轻量化） | 无 | 🟡 中期目标 |

---

## 5. 审计总结

**PicoClaw 是一个架构清晰、代码质量较高的成熟轻量级 AI 助手框架。**

核心接管重点（按优先级）：
1. **MVP 优先**：增强 SubAgent 编排系统（在现有 spawn 工具基础上扩展 Coordinator 模式）
2. **MVP 优先**：Android/Termux 适配文档 + 配置指南
3. **Phase 2**：可选云记忆体模块（设计接口先，实现后）
4. **Phase 3**：记忆分析工具 + Browser Control Skill

技术风险：
- 云记忆向量搜索需要 CGO（与纯 Go 优先原则冲突），需评估替代方案（如 `sqlite-vec` CGO 可选构建 tag）
- SubAgent 编排扩展需要谨慎设计以保持向后兼容
