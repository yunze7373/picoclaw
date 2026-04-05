# 📘 Phase/INBOX 添加规范 — Task Dispatcher 操作手册

> **使用者**: Task Dispatcher（控制层 AI）  
> **目的**: 将用户的文字/截图/视频翻译成高质量任务，注入永动引擎 INBOX  
> **重要性**: ⭐⭐⭐⭐⭐ 格式错误 = 引擎断链

---

## 🧠 你的工作流

```
用户输入（文字/截图/视频/参考图）
    ↓
你分析用户意图
    ↓
判断大小：
  小需求（1-3 改动）→ 直接写 INBOX 任务
  大需求（4+ 改动）→ 创建新 Phase
    ↓
写入文件 → 引擎自动拾取执行
```

## 🤖 Agent 路由表

写 INBOX 任务时，必须指定 Agent：

| 问题类型 | 指定 Agent | Pipeline |
|----------|-----------|----------|
| UI/视觉问题 | `senior-ui-ux-designer` | `ui_redesign` |
| 功能 bug | `root-cause-debugging-engineer` | `incident` |
| 新功能 | `principal-architect` → orchestrator | `new_feature` |
| 性能太慢 | `principal-performance-engineer` | — |
| 代码太乱 | `principal-refactoring-engineer` | `code_quality` |
| 缺少测试 | `software-test-engineer` | — |
| 缺少文档 | `technical-documentation-engineer` | — |
| 全面审查 | Phase R 流水线 | `phase_r_review` |

## 📋 约束协议：Task Injection Protocol (TIP)

你生成的每一条任务，必须绝对协议化，将具体的执行数据封装入 YAML 块中，以便执行引擎进行正则匹配与无损转化。

```markdown
❌ 差：
- [ ] 首页不好看

✅ 好：
- [ ] 🔴 **[P0] 首页 Hero 区视觉重塑**
  ```yaml
  tip_version: "1.0"
  agent_routing: "senior-ui-ux-designer"
  pipeline: "ui_redesign"
  impacted_files:
    - "src/pages/Home.tsx"
    - "src/styles/hero.css"
  context_analysis: "用户觉得背景纯白无层次，CTA按钮平淡"
  acceptance_criteria:
    - "渐变背景实现 (参考: Linear.app)"
    - "CTA 按钮渐变蓝加发光"
    - "主标题字体更换为 Inter 700 48px"
  references:
    - "https://linear.app"
  ```
```

---

## ⚠️ 常见错误（禁止这样做）

### ❌ 错误 1: 只添加 .md 文件，不创建 .json 计划
```
错误做法:
- 只创建了 plan_021_model_upgrade.md
- 没有创建 plan_021.json

后果: 系统无法解析任务列表，推进中断
```

### ❌ 错误 2: 只在 INBOX 添加条目，没有 Phase R Trigger
```
错误做法:
{
  "weeks": {
    "week112": { "goal": "Phase R 审查" }
  }
}
// 缺少 phase_r_trigger 配置

后果: Phase 完成后不会自动生成下一计划
```

### ❌ 错误 3: 忘记更新 TODOS.md 和 PROGRESS_STATUS.json
```
错误做法:
- 创建了 plan_021.json
- 但 TODOS.md 还是 Phase 16 的内容
- PROGRESS_STATUS.json 还是 phase20_completed

后果: AI 读取错误状态，执行混乱
```

### ❌ 错误 4: INBOX 条目没有指向正确的计划文件
```
错误做法:
- [ ] Phase 21
  - 读取：`.progress/plans/plan_021_model_upgrade.md`  ← .md 文件

正确做法:
- [x] Phase 21
  - 读取：`.progress/plans/plan_021.json`  ← .json 文件

后果: 系统无法解析任务结构
```

---

## ✅ 正确流程（必须按此顺序）

### 场景 A: 用户提出新需求，需要添加新 Phase

#### 步骤 1: 创建 .md 计划书（可选，用于详细设计）
```markdown
文件：.progress/plans/plan_XXX_description.md
内容：详细的设计文档、任务分解、参考链接
```

#### 步骤 2: 创建 .json 计划文件（必须）
```json
文件：.progress/plans/plan_XXX.json
```

**JSON 结构模板**:
```json
{
  "plan_id": "plan_025",
  "created": "2026-04-04",
  "phase": "Phase 25 - 描述性名称",
  "based_on": "INBOX.md user request + plan_025_description.md",
  "total_weeks": 6,
  "priority_order": ["P0-priority", "P1-secondary", "P2-tertiary"],
  "status": "pending",
  
  "phase_r_trigger": {
    "auto_execute": true,
    "next_plan": "plan_026",
    "tasks": [
      "Read .progress/PHASE_R_CHECKLIST.md to ensure strict transition",
      "Execute Phase R review & generation",
      "Execute Memory Archiving (Move [x] tasks to archive/)",
      "Call semantic-commit-generator to run 'git add -A && git commit'",
      "Check INBOX.md for new instructions",
      "Generate plan_026.json",
      "Continue executing plan_026 first task"
    ]
  },
  
  "weeks": {
    "week130": {
      "goal": "P0: 第一个周的目标",
      "priority": "P0-priority",
      "tasks": [
        {
          "id": "W130-1",
          "title": "任务标题",
          "type": "frontend|backend|test|docs|review|auto",
          "files": ["path/to/file.tsx"],
          "details": "任务详细描述"
        }
      ]
    },
    "week135": {
      "goal": "Phase R 审查 + 归档 + 下一计划生成",
      "priority": "phase-r",
      "tasks": [
        {
          "id": "W135-1",
          "title": "Phase R 审查、归档历史记忆、生成 plan_026.json",
          "type": "review",
          "files": [".progress/reports/AUDIT_REPORT_025.md", ".progress/plans/plan_026.json"],
          "details": "1. 执行审查。2. 将 INBOX.md 和 TODOS.md 中已完成的 [x] 项剪切到 .progress/archive/ 目录下归档。3. 生成新计划"
        },
        {
          "id": "W135-2",
          "title": "立即执行 Phase 26 第一个任务",
          "type": "auto",
          "details": "不询问，直接继续执行"
        }
      ]
    }
  }
}
```

#### 步骤 3: 更新 PROGRESS_STATUS.json
```json
{
  "project": "ShortDramaStudio",
  "current_phase": "Phase 25 - 描述性名称",
  "current_week": 130,
  "overall_progress": 97,
  "status": "phase25_in_progress",
  "current_plan": "plan_025.json",
  
  "phase25_描述性名称": {
    "status": "active",
    "plan": "plan_025.json",
    "current_week": "week130",
    "phase_r_trigger": {
      "auto_execute": true,
      "next_plan": "plan_026",
      "triggered": false
    }
  },
  
  "phase26_下一个计划": {
    "status": "pending",
    "plan": "plan_026.json",
    "phase_r_trigger": {
      "auto_execute": true,
      "next_plan": "plan_027",
      "triggered": false
    }
  },
  
  "last_updated": "2026-04-04",
  "auto_push_enabled": true,
  "phase_r_auto_execute": true
}
```

#### 步骤 4: 更新 TODOS.md
```markdown
# Phase 25 任务清单

> ⚠️ 无论你是什么模型，读到此文件必须执行以下规则：
> 1. 先读取 .progress/AI_INSTRUCTIONS.md
> 2. 完成当前任务后立即执行下一个 pending 任务
> 3. 所有任务完成后执行 Phase R → 生成新计划 → 继续执行
> 4. 永远不要问"是否继续"，直接执行

---

## ✅ Phase 24 已完成
[已完成的 Week 列表]

---

## ⏳ Phase 25 待执行 (plan_025.json)

### P0: 第一个周的任务 (Week 130)
- [ ] **W130-1**: 任务标题 ⏳
  - 类型：frontend | 文件：`path/to/file.tsx`
  - 详情：任务描述

### Phase R: 审查 + Phase 26 生成 (Week 135)
- [ ] **W135-1**: Phase R 审查 + 生成 plan_026.json ⏳
- [ ] **W135-2**: 立即执行 Phase 26 第一个任务 ⏳

---

## 🔒 Phase 26-XX 待启动
[后续 Phase 的简要描述]

---

**当前执行**: Phase 25 W130-1 → 任务描述  
**下一个**: W130-2 → 下一个任务  
**Phase R 触发**: Phase 25 全部完成后自动执行 → 生成 plan_026.json → 继续执行
```

#### 步骤 5: 更新 INBOX.md
```markdown
## 待处理

- [x] **Phase 24 完成**: Midjourney Standard UI Overhaul ✅
  - 读取：`.progress/plans/plan_024.json`
  - 状态：✅ 完成，Phase R 触发器已配置 → 自动生成 plan_025.json

- [x] **Phase 25 启动**: 新 Phase 名称 ✅
  - 读取：`.progress/plans/plan_025.json`
  - 核心任务：一句话描述
  - P0/P1/P2: 主要任务分类
  - 状态：✅ JSON 计划已创建，Phase R 触发器已配置

- [ ] **BUG-X**: 如果有 bug 待修复
  - 文件：`path/to/file.tsx`
  - 状态：待修复
```

---

## 🔑 关键检查清单

创建新 Phase 后，必须确认以下内容：

### 文件检查
- [ ] `.progress/plans/plan_XXX.json` 已创建
- [ ] `plan_XXX.json` 中包含 `phase_r_trigger` 配置
- [ ] `plan_XXX.json` 最后一周包含 "Phase R 审查 + 生成 plan_XXX+1.json" 任务
- [ ] `plan_XXX.json` 最后一周包含 "立即执行下一计划第一个任务" 任务

### 状态文件检查
- [ ] `PROGRESS_STATUS.json` 的 `current_phase` 已更新
- [ ] `PROGRESS_STATUS.json` 的 `status` 改为 `phaseXXX_in_progress`
- [ ] `PROGRESS_STATUS.json` 的 `current_plan` 指向新 plan
- [ ] `PROGRESS_STATUS.json` 中包含新 Phase 的配置块
- [ ] `PROGRESS_STATUS.json` 中设置了 `auto_push_enabled: true` 和 `phase_r_auto_execute: true`

### 任务文件检查
- [ ] `TODOS.md` 包含新 Phase 的详细任务列表
- [ ] `TODOS.md` 顶部有正确的执行规则说明
- [ ] `TODOS.md` 底部有 "当前执行/下一个/Phase R 触发" 指引

### INBOX 检查
- [ ] `INBOX.md` 中已完成 Phase 标记为 ✅
- [ ] `INBOX.md` 中新 Phase 指向 `.json` 文件（不是 .md）
- [ ] `INBOX.md` 中标注 "Phase R 触发器已配置"

---

## 📋 快速模板（复制即用）

### plan_XXX.json 最小可用模板
```json
{
  "plan_id": "plan_025",
  "created": "2026-04-04",
  "phase": "Phase 25 - 名称",
  "total_weeks": 6,
  "status": "pending",
  "phase_r_trigger": {
    "auto_execute": true,
    "next_plan": "plan_026",
    "tasks": [
      "Execute Phase R review",
      "Check INBOX.md",
      "Generate plan_026.json",
      "Continue executing plan_026 first task"
    ]
  },
  "weeks": {
    "week130": {
      "goal": "P0: 第一周目标",
      "tasks": [
        {
          "id": "W130-1",
          "title": "任务",
          "type": "backend",
          "files": ["file.py"],
          "details": "详情"
        }
      ]
    },
    "week135": {
      "goal": "Phase R + plan_026 生成",
      "tasks": [
        {
          "id": "W135-1",
          "title": "Phase R 审查 + 生成 plan_026.json",
          "type": "review",
          "files": [".progress/reports/AUDIT_REPORT_025.md", ".progress/plans/plan_026.json"]
        },
        {
          "id": "W135-2",
          "title": "立即执行 Phase 26 第一个任务",
          "type": "auto"
        }
      ]
    }
  }
}
```

### PROGRESS_STATUS.json 追加模板
```json
"phase25_名称": {
  "status": "active",
  "plan": "plan_025.json",
  "current_week": "week130",
  "phase_r_trigger": {
    "auto_execute": true,
    "next_plan": "plan_026",
    "triggered": false
  }
},
"phase26_下一个": {
  "status": "pending",
  "plan": "plan_026.json",
  "phase_r_trigger": {
    "auto_execute": true,
    "next_plan": "plan_027",
    "triggered": false
  }
}
```

---

## 🚨 常见故障排查

### 问题：Phase 完成后没有自动生成下一计划
**原因**: `phase_r_trigger` 配置缺失或 `type: "auto"` 任务不存在  
**解决**: 检查最后一周是否有 `"type": "auto"` 的任务

### 问题：AI 不执行新 Phase 的任务
**原因**: `PROGRESS_STATUS.json` 的 `status` 还是 `phaseXX_completed`  
**解决**: 改为 `phaseXX_in_progress`

### 问题：AI 读取错误的计划文件
**原因**: `INBOX.md` 指向 `.md` 而不是 `.json`  
**解决**: 改为 `.progress/plans/plan_XXX.json`

### 问题：TODOS.md 与计划不同步
**原因**: 创建计划后没有更新 TODOS.md  
**解决**: 每次创建计划后必须同步更新 TODOS.md

---

## 📞 总结

**核心规则**:
1. **必须创建 .json 计划** - .md 只是参考，.json 才是执行依据
2. **必须配置 phase_r_trigger** - 否则推进链条断裂
3. **必须更新 4 个文件** - plan_XXX.json + PROGRESS_STATUS.json + TODOS.md + INBOX.md
4. **最后一周必须有 auto 类型任务** - 确保 Phase R 后继续执行

**添加顺序**:
```
1. 创建 plan_XXX.json (含 phase_r_trigger)
   ↓
2. 更新 PROGRESS_STATUS.json (status → in_progress)
   ↓
3. 更新 TODOS.md (详细任务列表)
   ↓
4. 更新 INBOX.md (指向.json 文件)
   ↓
5. 验证检查清单
```

遵守此规范，自主推进系统永不断链！
