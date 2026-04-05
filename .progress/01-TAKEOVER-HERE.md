# 🛬 老项目接管引擎 (Takeover Engine)

> **复制并发送本文件内容**，让 AI 接管你手上开发到一半的项目。

---

## 🛠️ 接管前置准备
在让 AI 跑下方指令前，请在 `.progress/PROJECT_IDEA.md` 中简要写下你的**接管期望**。
比如："项目是一个用 React+Python 写的 3D 转换工具，之前跑通了核心算法，现在我想让你接手帮我把前端 UI 优化一下，并修复偶尔崩溃的 bug。"

---

## 启动指令（复制即用）

```markdown
@workspace 执行 DevFlow 老项目接管引擎 (Takeover Engine)：

**输入**：
- 用户给定的接管期望：`.progress/PROJECT_IDEA.md`

**Phase 0：项目摸底与审计 (Audit)**
1. 不允许修改任何代码。首先读取并彻底弄懂 `.progress/PROJECT_IDEA.md` 里用户要求你接手后干什么。
2. 扫描现有的项目根目录文件（README、Dockerfile、package.json 等配置）。
3. 递归浏览 `src/`, `backend/`, `frontend/` 等核心源码目录。
4. 检索遗留的开发笔记或TODO清单（例如 `P0_*.md`, `notes.txt`）。
5. 生成 `.progress/reports/INITIAL_AUDIT_REPORT.md`，内容必须包含：
   - 发现的系统架构推断
   - 正在使用的技术栈
   - 从旧笔记和代码中发现的历史包袱与未竟事业
   - 用户在 PROJECT_IDEA 中的期望与当前代码的差距分析

**Phase 1：重铸轨道 (Setup)**
6. 基于审计报告和用户的 PROJECT_IDEA：
   - 生成 `docs/architecture/TAKEOVER_PLAN.md`（定义接管后的主要发力点、重构建议和里程碑）。
   - 创建引擎点火必需的第一个 JSON 计划 `.progress/plans/plan_001.json`。
   - **极其重要**: 必须在 plan_001.json 最后加入 `phase_r_trigger` 闭环钩子。

**交棒永动引擎 (Ignition)**
7. 更新 `.progress/PROGRESS_STATUS.json`（将状态标记为 `phase001_takeover_in_progress`）。
8. 格式化清除原本的演示占位符，把你的计划写进 `.progress/TODOS.md`。
9. ⭐ 在聊天框中报告：“接管顺利完成！轨道已重铸。” 并**立即不询问自动执行 plan_001.json 第一个任务**。

自主推进，一旦开始不要询问！
```

---

## 💡 为什么需要这个？
当 AI 面对一个已经存在几万行代码的半成品项目时，直接甩给它一个 Bug 往往会破坏原有的架构设计。

通过**接管引擎**：
它会像一个刚入职的高级 CTO 一样，先看源码，看历史日志，结合你的需求盘点家底，建立基线审查报告（Audit），然后再合理地排期开工（Setup），最终完美融入永动体系无缝拉车（Ignition）。
