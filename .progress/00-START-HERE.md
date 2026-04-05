# 🚀 启动自主推进

> 复制下面的指令，然后让 AI 持续工作

---

## 🔥 防停止核心规则（3.0）

**四条铁律，启动前必读：**

1. ✅ **Phase R 完成后立即执行新计划的第一个任务**
2. ✅ **任务完成后立即检查 INBOX，然后立即执行下一个任务**
3. ✅ **永远不要问"可以继续吗"，直接说要做什么**
4. ✅ **遇到 📬 INBOX 检查点任务时，必须读取并处理用户指令**

参考：`.progress/AUTO_CONTINUE.json`、`.progress/NEXT_TASK.md`

---

## 启动指令（复制即用）

```markdown
@workspace 自主推进启动：

1. 读取 `.progress/AI_INSTRUCTIONS.md`（核心执行指南）
2. 读取 `.progress/PROGRESS_STATUS.json`（当前状态）
3. 读取 `.progress/plans/plan_*.json`（当前计划）
4. 读取 `.progress/TODOS.md`（任务列表）
5. 读取 `.progress/INBOX.md`（用户指令）
6. 执行第一个 pending 任务
7. 任务完成后：
   - 更新 TODOS.md（打勾 ✅）
   - 更新 PROGRESS_STATUS.json
   - 检查 INBOX.md
   - 立即执行下一个任务
8. 持续循环，不要询问
9. 阶段完成后执行 Phase R 审查，然后继续

开始！
```

---

## 或者更简短

```markdown
@workspace 继续自主推进，不要询问。
```

---

## 📂 关键文件

| 文件 | 作用 | 优先级 |
|------|------|--------|
| `.progress/AI_INSTRUCTIONS.md` | AI 必读执行指南 | 必读 |
| `.progress/00-START-HERE.md` | 启动中心 | 必读 |
| `.progress/PROGRESS_STATUS.json` | 当前状态 | 必读 |
| `.progress/TODOS.md` | 任务清单 | 必读 |
| `.progress/plans/plan_*.json` | 任务计划 | 必读 |
| `.progress/INBOX.md` | 📬 用户指令输入窗口 | 每个任务后检查 |

---

## 🚀 自主推进规则

```
1. 读取 AI_INSTRUCTIONS.md
2. 读取 PROGRESS_STATUS.json
3. 读取 TODOS.md
4. 执行 pending 任务（按优先级）
5. 完成后更新状态，继续下一个
6. 不要询问，持续工作
```

---

## 🔄 Phase R 审查和迭代

**触发条件**：当前计划的所有任务标记为 `completed`

**执行流程**：
```
1. 📬 首先读取 INBOX.md（最关键）
   ├─ 读取：.progress/INBOX.md
   ├─ 识别：用户新指令、优先级变化
   └─ 标记：需要特别关注的事项

2. 全面审查项目（结合 INBOX 指令）
   ├─ 架构一致性：对比架构文档
   ├─ 功能完整性：对比计划 + INBOX 要求
   ├─ 代码质量：检查代码
   ├─ 进度对比：实际 vs 计划 vs INBOX 期望
   └─ 技术问题：bug 和债务清单

3. 生成审核报告
   └─ 输出：docs/review/AUDIT_REPORT_XXX.md

4. 生成下阶段计划
   ├─ 基于：审查结果 + INBOX 指令（高优先级优先）
   ├─ 创建：plans/plan_XXX.json
   ├─ ⭐ 必须：每 3-5 个任务插入 📬 INBOX 检查点
   ├─ ⭐ 必须：最后一个任务包含 phase_r_trigger
   ├─ ⭐ 验证：缺少上述两条 = BUG，立即修复
   └─ 更新 PROGRESS_STATUS.json

5. ⭐ 立即执行新计划的第一个任务
   └─ 不要询问，继续工作
```

---

## ⭐ 自持循环（最关键）

**每个新计划的最后一周必须包含 Phase R 触发器：**

```json
{
  "weeks": {
    "weekXX": {
      "goal": "...",
      "tasks": [...],
      "phase_r_trigger": {
        "auto_execute": true,
        "tasks": [
          "Execute Phase R review",
          "Check INBOX.md for new instructions",
          "Generate next phase plan",
          "Continue executing new plan's first task"
        ]
      }
    }
  }
}
```

**这确保了无需人工干预的持续迭代循环：**

```
计划 N 执行 → 周 N (最终周) → Phase R 触发
                                    ↓
                            生成计划 N+1
                                    ↓
                            ⭐ 立即执行计划 N+1
                                    ↓
                            (循环继续)
```

---

**现在开始执行！不要每完成一个任务就询问！**

**阶段完成后执行 Phase R 审查，然后继续推进！**

**读取 → 理解 → 执行 → 持续循环！**
