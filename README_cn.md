# EinoRag

EinoRag 是一个基于字节跳动 Eino 框架的 RAG（检索增强生成，Retrieval-Augmented Generation）API SDK，帮助开发者轻松集成向量检索与生成式 AI 能力。项目采用 Go 1.22+，支持 Google Gemini Embedding，配置灵活，易于扩展。

---

## 目录结构

```
/Eino-rag      # RAG 核心功能实现（如向量上传、检索等）
/internal
  ├─ config    # 配置加载与管理（如 viper、yaml 解析等）
  ├─ embadding # 向量化相关逻辑
  └─ rag       # RAG 业务逻辑实现
/pkg
  ├─ wokerpool # 通用协程池
  └─ logger    # 日志组件
main.go        # 示例入口，演示上传与检索流程
README.md      # 项目说明文档
```

如需更详细的模块说明或接口文档，请告知具体需求！
---
## 项目架构

![EinoRag 架构图](images/rag.png)

上图展示了 EinoRag 的核心架构流程，包括数据上传、向量化、存储与检索等主要环节，便于开发者快速理解整体实现思路。

---
## 快速开始

1. **克隆项目并进入目录**
   ```bash
   git clone https://github.com/leebrouse/EinoRag.git
   cd EinoRag
   ```

2. **准备环境**
   - 安装 Go 1.22 及以上版本
   - 设置 Gemini API Key
     ```bash
     export GOOGLE_API_KEY="<你的_API_KEY>"
     ```
   - 编辑 `internal/config/global.yaml`，配置 embedding 模型
     ```yaml
     gemini:
       embedder: gemini-embedding-001
     ```

3. **运行示例**
   ```bash
   go run main.go
   ```
   main.go 会上传 PDF 文档并通过向量数据库检索内容。

4. **运行集成测试**
   ```bash
   go test ./test -run TestGeminiEmbedder_Real -v
   ```

---

## 主要功能

- 支持 PDF 文档向量化上传与检索
- 封装 Gemini Embedding API
- 灵活的配置管理（支持环境变量与 YAML 文件）
- 内置协程池与日志模块，便于扩展

---

## 常见问题

- 配置优先级：如 viper 未读取到环境变量，建议在 global.yaml 明确配置
- 鉴权失败：请确认 GOOGLE_API_KEY 已正确导出且有权限

---

## 参考

- [EinoRag GitHub 仓库](https://github.com/leebrouse/EinoRag)
- [Eino 文档](https://www.cloudwego.io/zh/docs/eino/quick_start/)
- [Google Gemini Embedding 文档](https://ai.google.dev/gemini-api/docs/migrate-to-cloud?hl=zh-cn)

---

如需详细 API 或二次开发文档，请参考各子目录源码及注释。
