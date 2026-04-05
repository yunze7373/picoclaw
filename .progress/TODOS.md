# 📋 TODOS - Phase 4: 多云向量嵌入 + 记忆检索质量

> ⚠️ **无论你是什么模型（Opus/GPT/Auto/Sonnet），读到此文件必须执行以下规则：**
> 1. **先读取 `.progress/AI_INSTRUCTIONS.md`** — 获取完整执行规则
> 2. **完成当前任务后，立即执行下一个 pending 任务**
> 3. **所有任务完成后，执行 Phase R → 生成新计划 → 继续执行**
> 4. **永远不要问"是否继续"或"要我继续吗"，直接执行**
> 5. **这是永动系统，任务永远不会"全部完成"，Phase R 会生成新任务**

> 🎯 目标：多云向量嵌入模型支持 + Supabase 真实向量搜索  
> 📊 进度：0/13 完成  
> �� 计划文件：`.progress/plans/plan_004.json`  
> 📬 INBOX 来源：Phase 4 P0 - 云记忆体多云嵌入

---

## ✅ 历史完成归档

- Phase 1 (9/9): SubAgent 编排 + Termux 适配
- Phase 2 (8/8): 云记忆体模块 + 分析工具  
- Phase 3 (6/6): 备份管理 + Engine API + E2E 接入

---

## 🔄 Phase 4 任务（当前）

### Week 1: EmbeddingProvider 接口层

- [ ] **task_024** [P0]: 设计 EmbeddingProvider 接口 + NoopProvider
- [ ] **task_025** [P0] 依赖 024: OpenAI Embeddings 后端
- [ ] **task_026** [P0] 依赖 024: Ollama 本地嵌入后端

### Week 2: 多云实现 + Supabase 集成

- [ ] **task_027** [P0] 依赖 024: Google Vertex AI / AI Studio 嵌入
- [ ] **task_028** [P0] 依赖 024: 阿里云百炼嵌入
- [ ] **task_029** [P1] 依赖 024: DeepSeek Embedding 后端
- [ ] **task_030** [P0] 依赖 025+026: 集成嵌入到 Supabase 后端

### Week 3: 缓存 + 配置 + 修复

- [ ] **task_031** [P1] 依赖 024: 嵌入缓存层 (LRU)
- [ ] **task_032** [P0] 依赖 024: EmbeddingConfig 配置系统
- [ ] **task_033** [P1]: 修复 GetHealth FTS5 副作用
- [ ] **task_034** [P1] 依赖 030+031+032: 集成测试 + 文档
- [ ] **task_035**: 📬 INBOX 检查点
- [ ] **task_036**: 🔥 Phase R → plan_005.json → 立即执行

---

## 📌 立即执行

**task_024** (无依赖，P0): 设计 EmbeddingProvider 接口

---

**开始工作！持续执行！**
