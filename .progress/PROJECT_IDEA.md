# 项目想法

> 填写你的项目创意，AI 会基于此搜索参考项目并设计架构
> 不需要很详细，有大致方向即可

---

## 要做什么

为 PicoClaw 项目贡献新功能，作为 contributor 增强其多 Agent 协作能力和可选的云记忆体系统。重点借鉴 openclaw-termux 的 Android 轻量化部署经验，确保所有新功能在小型设备（尤其是 Android/Termux 环境）上可运行。

---

## 核心功能（1-5 个）

1. **增强的 SubAgent 编排系统** - 借鉴 claude-code 的多 Agent 并行协作架构，在现有 SubTurn 基础上实现 coordinator-worker 模式，支持 Research/Synthesis/Implementation/Verification 四阶段工作流

2. **可选的云记忆体扩展** - 参考 `han-skills/memory-manager` 的 Supabase+pgvector 方案，设计可选的云记忆存储模块：
   - 默认关闭，不影响现有 SQLite 本地记忆系统
   - 启用时提供向量语义搜索、跨设备记忆同步
   - 完全可选，开启或不开启对核心功能无任何影响

3. **Android/Termux 优化** - 参考 `yunze7373/openclaw-termux` 的部署经验：
   - 针对 Android 文件系统适配
   - ARM64 原生模块编译支持
   - 资源占用优化（目标 <50MB RAM）

4. **记忆分析工具** - 基于现有 LCM（Short-Term Memory Engine）添加：
   - 记忆健康状态检查
   - 简单的统计面板（记忆数量、Token 使用）
   - 可选的定时备份 Cron Job

5. **Browser Control 轻量化集成** - 参考 OpenClaw 的 CDP 控制，但针对小型设备优化：
   - 作为可选 Skill 而非核心功能
   - 支持基本的网页截图、DOM 查询
   - 不依赖完整 Chrome，支持无头模式

---

## 目标用户

- PicoClaw 社区贡献者和维护者
- 在 Android 设备、RISC-V 开发板等低成本硬件上运行 AI Agent 的开发者
- 需要轻量级、可选云同步记忆功能的用户

---

## 不做什么（边界）

### 明确不做的功能

- **不破坏轻量化特性** - 所有新功能必须是可选模块，默认关闭时不增加资源占用
- **不强制云依赖** - 云记忆体完全可选，本地 SQLite 方案仍是默认
- **不做平台特定客户端** - macOS menu bar app、iOS/Android 原生 nodes
- **不重写现有核心** - 基于现有 SubTurn、LCM、MCP 系统扩展，不推翻重来
- **不做加密货币、NFT 等无关功能**
- **不牺牲 RISC-V/ARM/MIPS 架构支持** - 保持跨平台兼容性

### 资源约束红线

| 指标 | 当前 | 目标上限 |
|------|------|----------|
| 基础 RAM | <10MB | <50MB（启用所有可选功能） |
| Android 支持 | 已有 | 保持 Termux 兼容 |
| 启动时间 | <1s | <5s（即使启用云记忆体） |

---

## 搜索关键词

- multi-agent orchestration lightweight
- Android Termux AI agent deployment
- SQLite vector search pgvector alternative
- Go memory store embedded database
- coordinator worker pattern AI
- cross-device memory sync
- low resource AI assistant

---

## 我的偏好和约束

### 技术偏好

- **后端**：Go 1.25+（PicoClaw 主语言）
- **数据库**：SQLite（默认）+ 可选 Supabase/pgvector
- **部署**：Docker 可选，优先单二进制分发
- **架构**：模块化设计，功能可按需启用/禁用

### 时间约束

- **MVP（2-4 周）**：SubAgent 编排增强 + Android 适配文档
- **Phase 2（1-2 月）**：云记忆体可选模块
- **Phase 3（2-3 月）**：记忆分析工具 + Browser Control

### 资源约束

- 本地开发环境 + 远程测试服务器
- Android 设备测试（Termux 环境）
- 优先使用开源库和 PicoClaw 现有基础设施

### 必须支持

- **中文界面和文档**
- **向后兼容** - 不影响现有配置和功能
- **可选性** - 所有新功能可独立启用/禁用
- **跨平台** - Linux/Android/macOS/Windows + RISC-V/ARM/MIPS/x86

---

## 已有的想法（可选）

### PicoClaw 现有记忆系统分析

当前 PicoClaw 使用三层记忆架构：

```
┌─────────────────────────────────────────────────────┐
│  Layer 1: Short-Term (Seahorse LCM Engine)          │
│  - SQLite 后端，WAL 模式                              │
│  - 支持预算感知上下文装配 (Assembler)                 │
│  - LLM 驱动压缩 (Compaction)                          │
│  - 分片锁并发控制 (256 shards)                        │
├─────────────────────────────────────────────────────┤
│  Layer 2: Long-Term (MemoryStore)                   │
│  - 文件存储：memory/MEMORY.md                        │
│  - 每日笔记：memory/YYYYMM/YYYYMMDD.md               │
│  - 最近 N 天回溯                                     │
├─────────────────────────────────────────────────────┤
│  Layer 3: Session Store (JSONL)                     │
│  - 会话历史：*.jsonl 文件                             │
│  - 支持摘要存储 (SetSummary/GetSummary)              │
└─────────────────────────────────────────────────────┘
```

### 云记忆体扩展设计

**设计原则**：完全可选，零侵入

```
┌─────────────────────────────────────────────────────┐
│  可选云记忆体模块 (Cloud Memory Extension)          │
│  ┌───────────────────────────────────────────────┐  │
│  │ 配置开关：tools.memory.cloud.enabled          │  │
│  │ 后端：Supabase + pgvector                     │  │
│  │ 嵌入模型：DeepSeek/Gemini/Ollama (可选)       │  │
│  └───────────────────────────────────────────────┘  │
│                    ↓ 可选启用                        │
│  ┌───────────────────────────────────────────────┐  │
│  │ 功能：                                         │  │
│  │ - 向量语义搜索 (similarity search)            │  │
│  │ - 跨设备记忆同步                               │  │
│  │ - 定时备份 (Cron Job)                         │  │
│  │ - 记忆分析面板 (统计/健康检查)                 │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
                    ↓ 默认关闭时
┌─────────────────────────────────────────────────────┐
│  本地 SQLite 方案 (默认，无外部依赖)                 │
│  - 现有 LCM Engine 保持不变                         │
│  - 无网络请求，无 API 费用                            │
│  - 完全离线可用                                     │
└─────────────────────────────────────────────────────┘
```

### SubAgent 增强设计

扩展现有 `pkg/agent/subturn.go`：

```go
// 新增工具
type CoordinatorTools struct {
    SpawnWorker     // 并行启动 worker agent
    TeamCreate      // 创建 agent 团队
    SendMessage     // 跨 agent 通信
    QueryStatus     // 查询子 agent 状态
}

// XML 消息格式（参考 claude-code）
// <task-notification>
//   <task_id>xxx</task_id>
//   <phase>research|synthesis|implementation|verification</phase>
//   <payload>...</payload>
// </task-notification>
```

### Android/Termux 适配要点

参考 `yunze7373/openclaw-termux`：

1. **路径处理** - Termux 主目录 `$PREFIX/home` 适配
2. **原生模块** - sqlite-vec 等从源码编译（arm64）
3. **进程管理** - 使用 PM2 或原生 shell 脚本替代 systemd
4. **权限处理** - 无需 root，适配 Android 沙盒

---

## 参考项目（可选）

- https://github.com/sipeed/picoclaw - PicoClaw 主项目
- https://github.com/yunze7373/openclaw-termux - Android/Termux 部署优化参考
- https://github.com/yunze7373/claude-code - 多 Agent 编排参考
- https://github.com/han-skills/memory-manager - 云记忆体方案参考（Supabase+pgvector）
- https://github.com/openclaw/openclaw - Canvas/Browser Control 参考

---

> **填写完成后**，使用 ARCH_START.md 中的启动指令启动架构引擎
