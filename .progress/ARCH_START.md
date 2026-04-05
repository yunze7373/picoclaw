# 🏗️ 架构引擎启动

> 复制下面的指令，让 AI 从零开始构建项目架构

---

## 🔥 架构引擎核心规则

**三条铁律：**

1. ✅ **Phase 0 完成后立即进入 Phase 1，不要询问**
2. ✅ **Phase 1 完成后立即交棒永动引擎，不要停止**
3. ✅ **每一步都参考顶级开源项目，站在巨人肩膀上**

---

## 启动指令（复制即用）

```markdown
@workspace 执行 DevFlow 架构引擎：

**输入**：
- 项目想法：`.progress/PROJECT_IDEA.md`
- 架构指令：`.progress/ARCH_INSTRUCTIONS.md`

**Phase 0：自动研究**
1. 读取 PROJECT_IDEA.md，理解项目目标和约束
2. 基于项目类型搜索 GitHub 10-15 个顶级参考项目
3. 克隆到 references/ 目录
4. 用 ANALYSIS_TEMPLATE.md 逐个分析每个项目：
   - 架构设计、技术栈、Agent 设计、数据模型
   - 可借鉴的设计 + 不适合的点
5. 生成分析报告到 docs/analysis/
6. 总结共同模式，提取最佳实践

**Phase 1：架构设计**
7. 基于分析结果 + PROJECT_IDEA.md 的偏好约束：
   - 生成 docs/architecture/ARCHITECTURE.md（完整架构）
   - 生成 docs/architecture/RECOMMENDATIONS.md（借鉴建议）
   - 生成 docs/architecture/PHASE_PLAN.md（实施计划）
8. 设计数据模型、API、Agent 系统
9. 创建项目目录骨架（src/、tests/ 等）
10. 创建 VS Code Workspace 配置
11. 生成第一个任务计划 .progress/plans/plan_001.json

**交棒永动引擎**
12. 更新 .progress/PROGRESS_STATUS.json
13. 更新 .progress/TODOS.md
14. ⭐ 立即启动永动引擎，执行 plan_001 第一个任务

自主推进，不要询问！开始！
```

---

## 或者更简短

```markdown
@workspace 读取 .progress/PROJECT_IDEA.md 和 .progress/ARCH_INSTRUCTIONS.md，
执行架构引擎（Phase 0 研究 → Phase 1 设计 → 交棒永动引擎），
不要询问！
```

---

## 📂 架构引擎需要的文件

| 文件 | 作用 | 谁准备 |
|------|------|--------|
| `.progress/PROJECT_IDEA.md` | 项目创意、目标、偏好约束 | **用户填写** |
| `.progress/ARCH_INSTRUCTIONS.md` | 架构引擎执行规则 | 模板自带 |
| `.progress/ANALYSIS_TEMPLATE.md` | 参考项目分析模板 | 模板自带 |
| `.progress/ARCH_START.md` | 本文件（启动指令） | 模板自带 |

---

## 🏗️ 架构引擎执行流程

```
Phase 0：自动研究（2-3 小时）
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
读取 PROJECT_IDEA.md
    ↓
搜索 GitHub（10-15 个项目）
    ↓
克隆 → references/ref-*/
    ↓
逐个分析（用 ANALYSIS_TEMPLATE.md）
    ↓
生成 docs/analysis/ref-*.md
    ↓
总结共同模式 → docs/analysis/SUMMARY.md

Phase 1：架构设计（1-2 小时）
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
总结分析结果 + 用户偏好
    ↓
设计架构 → docs/architecture/ARCHITECTURE.md
    ↓
借鉴建议 → docs/architecture/RECOMMENDATIONS.md
    ↓
实施计划 → docs/architecture/PHASE_PLAN.md
    ↓
创建项目骨架（src/、tests/ 等）
    ↓
生成 plan_001.json + TODOS.md

交棒永动引擎（立即执行）
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
更新 PROGRESS_STATUS.json
    ↓
⭐ 执行 plan_001 第一个任务
    ↓
永动引擎接管 → 无限循环
```

---

## 预期输出

### Phase 0 输出

```
references/                     ← 参考项目源码
├── ref-project-a/
├── ref-project-b/
├── ref-project-c/
...

docs/analysis/                  ← 分析报告
├── ref-project-a.md
├── ref-project-b.md
├── ref-project-c.md
└── SUMMARY.md                  ← 共同模式总结
```

### Phase 1 输出

```
docs/architecture/              ← 架构设计
├── ARCHITECTURE.md             ← 完整架构设计
├── RECOMMENDATIONS.md          ← 借鉴建议
└── PHASE_PLAN.md               ← 实施阶段计划

src/                            ← 项目骨架
├── backend/
├── frontend/
└── agents/

.progress/plans/plan_001.json   ← 第一个执行计划
```

---

## ⭐ 架构引擎 → 永动引擎衔接

```
架构引擎完成 Phase 1
    ↓
自动生成 plan_001.json（含 📬 INBOX 检查点 + Phase R 触发器）
    ↓
自动更新 PROGRESS_STATUS.json
    ↓
⭐ 立即执行 plan_001 第一个任务（不要停！不要问！）
    ↓
永动引擎接管 → 无限循环 ♾️
```

**衔接铁律**：架构引擎的最后一步 = 永动引擎的第一步。中间不停顿。

---

**现在开始！读取 PROJECT_IDEA.md → 研究 → 设计 → 执行 → 永不停歇！**
