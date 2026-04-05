# Phase R — 阶段审查与迭代中心

> ⭐ **这是 Phase R 审查的核心入口**  
> 当阶段任务完成后，AI 必须读取本文件并执行审查流程

---

## ⭐ 核心执行指令（AI 必读）

**当阶段任务完成后，你必须严格按以下顺序执行：**

```
⚠️ 警告：以下步骤必须严格按顺序执行，跳过步骤 1 是严重错误！

步骤 1️⃣ → 📬 首先读取 .progress/INBOX.md
         这是 Phase R 的第一步，不读取 INBOX 不得继续！
         
步骤 2️⃣ → 读取 .progress/PHASE_R_CHECKLIST.md — 执行系统防呆自检与脱水操作
         (强制清理旧任务 [x] 到 archive，并执行 Git Commit 保存阶段锚点)

步骤 3️⃣ → 执行 Phase R 审查流程（使用多 Agent 专家审查，结合 INBOX 指令）

步骤 4️⃣ → 生成审核报告 docs/review/AUDIT_REPORT_XXX.md

步骤 5️⃣ → 生成下阶段计划（必须包含 phase_r_trigger）

步骤 6️⃣ → ⭐ 立即执行新计划的第一个 P0 任务

步骤 7️⃣ → 持续工作，不要询问
```

**关键：Phase R 的第一步必须是读取 INBOX.md！最后一环必须执行新计划！跳过 = 系统故障！**

---

## 📋 启动条件

当满足以下任一条件时，触发 Phase R：

1. `.progress/PROGRESS_STATUS.json` 中当前阶段的所有任务标记为 `completed`
2. `.progress/plans/plan_*.json` 中的 `phase_r_trigger` 被激活
3. 用户在 `INBOX.md` 中写入 "启动 Phase R 审查"

---

## 🔄 Phase R 执行流程

```
Phase R: 审查和迭代
│
├─ 步骤 1: 📬 首先读取 INBOX.md（最关键）
│   ├─ 读取：.progress/INBOX.md
│   ├─ 识别：用户新指令、反馈、想法
│   ├─ 标记：需要特别关注的事项
│   └─ 优先级：高优先级指令必须纳入新计划
│
├─ 步骤 2: 脱水归档与项目审查（结合 INBOX 指令，使用多 Agent 专家审查）
│   ├─ 🧹 执行归档：移动完毕的 [x] 任务至 .progress/archive/
│   ├─ 💾 代码存档：生成本阶段完成的 Git Commit
│   ├─ 🤖 多 Agent 专家审查（依次读取 Agent 人设文件并切换思维模式）：
│   │   ├─ 读取 .dev-flow/agents/staff-code-reviewer.md → 审查代码质量
│   │   ├─ 读取 .dev-flow/agents/principal-architect.md → 审查架构一致性
│   │   ├─ 读取 .dev-flow/agents/software-test-engineer.md → 审查测试覆盖
│   │   └─ 读取 .dev-flow/agents/technical-documentation-engineer.md → 审查文档
│   ├─ 功能完整性：对比计划 + INBOX 要求
│   ├─ 进度对比：实际完成 vs 计划 vs INBOX 期望
│   └─ 参考项目借鉴：检查 references/ 或 docs/review/ 中的借鉴记录
│
├─ 步骤 3: 生成审核报告
│   └─ 输出：docs/review/AUDIT_REPORT_XXX.md
│   └─ 包含：INBOX 指令处理状态
│
├─ 步骤 4: 生成下阶段计划
│   ├─ 基于审核报告中的改进建议
│   ├─ 结合：INBOX.md 中的新指令（高优先级优先）
│   ├─ 创建：.progress/plans/plan_XXX.json
│   ├─ ⭐ 必须：最后一周添加 phase_r_trigger
│   ├─ 更新：.progress/PROGRESS_STATUS.json
│   ├─ 更新：.progress/TODOS.md
│   ├─ ⭐ **设置 AUTO_CONTINUE.json** → `auto_continue_enabled: true`
│   └─ ⭐ **标记下一个任务** → `.progress/NEXT_TASK.md`
│
└─ 🔥 步骤 5: **立即执行新计划的第一个任务**（不要停！）
    ├─ 检查 AUTO_CONTINUE.json → auto_execute
    ├─ 如果 auto_execute == true → **直接开始**第一个 P0 任务
    ├─ ❌ 不要说："如果你愿意，可以继续..."
    ├─ ✅ 直接说："现在执行 [任务名]..."
    └─ 开始代码执行，不要停顿
```

---

## 📝 审核报告模板

```markdown
# 项目审核报告 #{编号}

## 一、审查总览
[评分表格]

## 二、架构一致性检查
[符合/偏离清单]

## 三、功能完整性检查
[完成/未完成清单]

## 四、代码质量检查
[优点/改进清单]

## 五、技术问题清单
[Bug 清单、技术债务]

## 六、参考项目借鉴审查（新增）

### 已借鉴的实现
| 任务 | 参考项目 | 借鉴内容 | 适配情况 |
|------|----------|----------|----------|
| task_009 | ref-openOii | ModelScope API 集成 | ✅ 适配良好 |

### 可借鉴但未利用
| 参考项目 | 文件/设计 | 可借鉴点 | 建议 |
|----------|----------|----------|------|
| ref-jellyfish | tasks.py | Task 模式 | 建议下阶段引入 |

### 借鉴来源记录
- ✅ 代码注释中已注明借鉴来源
- ✅ 提交信息中已感谢原作者
- ⚠️ 部分实现缺少 LICENSE 合规检查

## 七、改进建议
P0 必须完成：...
P1 应该完成：...
P2 可以完成：...

## 八、下阶段建议
[基于审查的建议]
```

---

## 📝 下阶段计划模板（必须包含 phase_r_trigger）

```json
{
  "plan_id": "plan_XXX_phaseY",
  "weeks": {
    "weekXX": {
      "goal": "最终周目标",
      "tasks": [
        {...},
        {
          "id": "WXX-LAST",
          "title": "🔥 Phase R 审查 + 生成下一阶段计划",
          "type": "meta",
          "priority": "critical",
          "auto_trigger": true,
          "details": "本阶段所有任务完成后自动执行：1) 生成审核报告 2) 检查 INBOX.md 3) 创建 plan_XXX.json 4) 立即执行新计划第一个任务，不要询问"
        }
      ],
      "phase_r_trigger": {
        "auto_execute": true,
        "tasks": [
          "Read .progress/PHASE_R_CHECKLIST.md to ensure strict transition",
          "Execute Phase R review",
          "Execute Memory Archiving (Move [x] tasks to archive/)",
          "Call semantic-commit-generator to run 'git add -A && git commit'",
          "Check INBOX.md for new instructions",
          "Generate next phase plan (plan_XXX.json)",
          "Continue executing new plan's first task"
        ]
      }
    }
  }
}
```

---

## 🚀 手动启动 Phase R

如果你需要手动启动 Phase R，请向 AI 发送：

```
请读取 .progress/PHASE_R_START.md 启动 Phase R 审查
```

然后 AI 会自动执行上述流程。

---

## ⭐ 自持循环（Critical）

**Phase R 的最终目标不是生成报告，而是：**
1. 完成记忆清理与存档
2. 生成下阶段计划
3. **立即开始执行新计划**

这确保了无需人工干预的持续迭代循环：

```
Phase N 执行 → Phase N 完成 → Phase R 审查 (含清理与 Git 存档)
                                    ↓
                            Phase N+1 计划生成
                                    ↓
                            ⭐ 立即执行 Phase N+1
                                    ↓
                            Phase N+1 完成 → Phase R 审查
                                    ↓
                            (循环继续)
```

---

**现在执行 Phase R 审查！审查完成后立即继续执行新计划！**

**不要询问，持续工作！**
