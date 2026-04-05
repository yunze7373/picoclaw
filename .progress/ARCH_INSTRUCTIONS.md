# 🏗️ 架构引擎 AI 执行指南

> AI 必读！这是架构引擎的核心执行规则

---

## 🔥 核心规则

**五条铁律：**

1. ✅ **读取 PROJECT_IDEA.md 后立即开始搜索参考项目，不要询问**
2. ✅ **分析完所有参考项目后立即开始架构设计，不要停顿**
3. ✅ **架构设计完成后立即生成计划并交棒永动引擎**
4. ✅ **全程参考顶级开源项目，站在巨人肩膀上开发**
5. ✅ **委派专家 Agent 执行专业任务，不要笼统处理**

---

## 🤖 Agents 集成

架构引擎通过 Orchestrator 调度以下 Agent：

| 阶段 | Agent | 职责 |
|------|-------|------|
| Phase 0 分析 | `principal-architect` | 分析参考项目架构、提取设计模式 |
| Phase 1 设计 | `principal-architect` | 设计系统架构、模块边界、数据流 |
| Phase 1 文档 | `technical-documentation-engineer` | 生成架构文档、API 文档 |
| Phase 1 计划 | `agent-orchestrator` | 拆解任务并分配 agent |
| 交棒 | `semantic-commit-generator` | 提交架构设计成果 |

---

## 执行流程

### Phase 0：自动研究

```
步骤 1: 读取项目想法
━━━━━━━━━━━━━━━━━━
- 读取 .progress/PROJECT_IDEA.md
- 提取：项目类型、核心功能、目标用户、技术偏好、约束
- 提取：用户提供的搜索关键词

步骤 2: 搜索参考项目
━━━━━━━━━━━━━━━━━━
- 基于项目类型和关键词搜索 GitHub
- 选出 10-15 个高价值项目（star数 + 代码质量 + 活跃度）
- 优先选择最近 2 年活跃的项目
- 覆盖不同技术方向

步骤 3: 克隆参考项目
━━━━━━━━━━━━━━━━━━
- 克隆到 references/ref-{项目名}/
- 如果无法克隆，记录 URL 并直接在线分析

步骤 4: 逐个分析（委派 principal-architect）
━━━━━━━━━━━━━━━━━━
- 调用 principal-architect 分析每个项目的架构
- 用 ANALYSIS_TEMPLATE.md 的格式生成报告
- 输出到 docs/analysis/ref-{项目名}.md
- 重点关注：架构设计、技术栈、Agent 设计、数据模型
- 记录"可借鉴" + "不适合" 的点

步骤 5: 总结共同模式（principal-architect 汇总）
━━━━━━━━━━━━━━━━━━
- principal-architect 对比所有参考项目
- 提取共同的架构模式
- 总结技术栈选择趋势
- 输出到 docs/analysis/SUMMARY.md
```

### Phase 1：架构设计

```
步骤 6: 设计系统架构（principal-architect 主导）
━━━━━━━━━━━━━━━━━━
- 委派 principal-architect 基于分析结果 + PROJECT_IDEA.md
- 输出：docs/architecture/ARCHITECTURE.md
- 包含：
  ├─ 系统概述（目标、功能、用户）
  ├─ 架构层级图
  │   ├── Agent 层
  │   ├── 应用层
  │   ├── 核心层
  │   ├── 数据层
  │   └── 基础设施层
  ├─ 技术栈选择（附理由）
  ├─ 核心模块设计（职责、接口、依赖）
  ├─ 数据模型设计（实体、关系）
  ├─ Agent 设计（类型、职责、协作）
  ├─ API 设计（端点、协议）
  └─ 实施计划（分阶段）

步骤 7: 生成借鉴建议（technical-documentation-engineer）
━━━━━━━━━━━━━━━━━━
- 输出：docs/architecture/RECOMMENDATIONS.md
- 每个借鉴点标注来源项目
- 说明如何适配到我们的架构

步骤 8: 生成实施计划
━━━━━━━━━━━━━━━━━━
- 输出：docs/architecture/PHASE_PLAN.md
- 拆解为 Phase 1 / Phase 2 / Phase 3
- 每个 Phase 有明确的里程碑和交付物

步骤 9: 创建项目骨架
━━━━━━━━━━━━━━━━━━
- 创建目录结构（src/、tests/ 等）
- 创建 VS Code Workspace 配置
- 初始化基础配置文件

步骤 10: 生成第一个执行计划（agent-orchestrator 拆解）
━━━━━━━━━━━━━━━━━━━━━━━━
- 创建 .progress/plans/plan_001.json
- agent-orchestrator 基于 PHASE_PLAN.md 第一阶段拆解具体任务
- ⭐ 必须：每个任务指定 agent 和 pipeline
- ⭐ 必须：每 3-5 个任务插入 📬 INBOX 检查点
- ⭐ 必须：最后一个任务包含 phase_r_trigger
- 更新 PROGRESS_STATUS.json
- 更新 TODOS.md
- semantic-commit-generator 提交架构设计成果
```

### 交棒永动引擎

```
步骤 11: 衔接
━━━━━━━━━━━━
- ⭐ 立即执行 plan_001 的第一个任务
- 永动引擎接管
- 禁止询问"是否可以继续"
```

---

## 参考项目分析规则

### 搜索策略

| 渠道 | 用途 |
|------|------|
| GitHub Topics | 发现同类项目标签 |
| GitHub Trending | 发现热门项目 |
| GitHub Search | 关键词精确搜索 |
| Google | 发现非 GitHub 项目 |
| Papers With Code | 发现学术项目 |

### 选择标准

- ✅ Star 数 > 100（有社区认可）
- ✅ 最近 6 个月有更新（活跃维护）
- ✅ 有清晰的目录结构（代码质量好）
- ✅ 技术栈与用户偏好兼容
- ❌ 排除纯 Demo / 教程项目
- ❌ 排除许可证不兼容的项目

### 分析重点

| 维度 | 说明 | 重要性 |
|------|------|--------|
| 架构设计 | 分层、模块划分、设计模式 | ⭐⭐⭐⭐⭐ |
| 技术栈 | 框架、库、基础设施 | ⭐⭐⭐⭐ |
| Agent 设计 | Agent 类型、职责、协作 | ⭐⭐⭐⭐ |
| 数据模型 | 核心实体、关系设计 | ⭐⭐⭐⭐ |
| API 设计 | 端点、协议、认证 | ⭐⭐⭐ |
| 目录结构 | 组织方式、命名规范 | ⭐⭐⭐ |
| 部署方式 | Docker、CI/CD | ⭐⭐ |

---

## 架构设计规则

### ✅ 必须做

1. **先分析再设计** — 至少分析 10 个参考项目
2. **分层设计** — 职责清晰，层间解耦
3. **尊重用户偏好** — PROJECT_IDEA.md 中的技术偏好优先
4. **记录决策依据** — 每个技术选择注明"为什么"和"参考来源"
5. **预留扩展点** — 但不要过度设计

### ❌ 禁止做

1. ❌ 不分析就开始设计
2. ❌ 只看一个参考项目就定架构
3. ❌ 忽视用户的技术偏好和约束
4. ❌ 不记录借鉴来源
5. ❌ 边做边改架构（Phase 0 定好再执行）

---

## 开发过程中的持续借鉴

架构引擎不仅在 Phase 0-1 工作，在**永动引擎执行期间**也持续发挥作用：

### 任务执行时

```
执行 task_009 - 实现图片生成服务
    ↓
先查看参考项目中类似的实现：
  - references/ref-openOii/backend/app/services/image.py
  - references/ref-jellyfish/.../image_generation_tasks.py
    ↓
分析借鉴点 → 实现我们的版本 → 记录借鉴来源
```

### 架构决策时

```
需要设计 Agent 系统
    ↓
查阅 docs/analysis/ 中各项目的 Agent 设计
    ↓
对比共同模式 → 设计我们的 Agent → 记录决策依据
```

### 遇到问题时

```
图片一致性问题
    ↓
搜索 references/ 中的解决方案
    ↓
学习 → 尝试 → 记录最佳实践
```

---

## 借鉴记录规范

### 代码注释

```python
# 图片生成服务
# 借鉴自：
# - ref-openOii: backend/app/services/image.py（重试机制）
# - ref-jellyfish: .../image_generation_tasks.py（供应商抽象）
# 修改：简化配置管理，适配我们的数据模型
```

### plan.json 中的 references 字段

```json
{
  "id": "task_009",
  "name": "实现图片生成服务",
  "references": [
    {
      "project": "ref-openOii",
      "file": "backend/app/services/image.py",
      "what_to_learn": "ModelScope API 集成和重试机制"
    }
  ]
}
```

---

**执行顺序：读取 PROJECT_IDEA → 搜索 → 分析 → 设计 → 生成计划 → 交棒永动引擎 → 永不停歇！**
