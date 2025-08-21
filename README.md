````markdown
````
# EinoRag
EinoRag 是一个基于 ByteDance Eino 框架之上的 RAG（Retrieval-Augmented Generation）API SDK，旨在帮助开发者更轻松地集成检索式生成能力。它使用了 Google Gemini 的 Embedding 接口，由 `viper` 负责配置加载与管理。  

---

## Table of Contents

- [特性](#特性)  
- [快速开始](#快速开始)  
- [配置说明](#配置说明)  
- [常见问题](#常见问题)   

---

## 特性

- 基于 Go 1.22+ 构建，推荐使用最新版本  
- 使用 `viper` 加载 YAML 配置及环境变量  
- 支持 Gemini Embedding 接口的封装与调用  
- 包含集成测试示例，便于验证配置与调用流程 

---
````
## 快速开始

1. **克隆项目**

   ```bash
   git clone https://github.com/leebrouse/EinoRag.git
   cd EinoRag
   go run main.go  #这个是一个 测试例子
````
2. **准备环境**

   * 安装 Go（版本 ≥ 1.22）

   * 设置 Gemini API Key：

     ```bash
     export GOOGLE_API_KEY="<你的_API_KEY>"
     ```

   * 编辑 `internal/config/global.yaml`，添加 embedding 模型：

     ```yaml
     gemini:
       embedder: gemini-embedding-001
     ```

3. **运行集成测试**

   ```bash
   go test ./test -run TestGeminiEmbedder_Real -v
   ```

   该测试会调用 Gemini 的实际嵌入接口，用于验证整个调用流程是否正确 ([GitHub][1])。

---
## 配置说明

* **`internal/config/global.yaml`**
  用于定义 embedding 模型字段，例如：

  ```yaml
  gemini:
    embedder: gemini-embedding-001
  ```

* **环境变量**
  SDK 支持通过环境变量读取 API Key（默认变量名为 `GOOGLE_API_KEY`）。如果你使用其他名称，可以参考 `google.golang.org/genai` 的文档进行配置 ([GitHub][1])。

---
* `viper.go`：负责全局配置与环境变量读取
* `gemini.go`：封装对 Gemini Embedding API 的调用
* `embedder_test.go`：提供如何调用与测试嵌入的参考代码 ([GitHub][1])

---

## 常见问题

* **配置问题**：若 `viper` 无法正确读取到环境变量中的 `gemini.embedder`，请优先在 `global.yaml` 中显式配置。当前实现对文件配置优先级较高 ([GitHub][1])。

* **鉴权错误**：若集成测试报告鉴权失败，请检查 `GOOGLE_API_KEY` 是否已成功导出，且具备访问 Gemini 的权限 ([GitHub][1])。

---
[1]: https://github.com/leebrouse/EinoRag "GitHub - leebrouse/EinoRag: EinoRag is a RAG API SDK built on top of the ByteDance Eino framework."
