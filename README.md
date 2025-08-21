## Eino

一个最小可用的示例，演示如何使用 Gemini 的 Embedding 能力并通过 `viper` 加载配置。

### 运行环境
- **Go**: 1.22+（建议）
- **依赖**:
  - `google.golang.org/genai`
  - `github.com/cloudwego/eino`
  - `github.com/spf13/viper`

### 配置
本项目使用 `viper` 在初始化时读取 `internal/config/global.yaml`，同时会自动读取环境变量（用于 API Key）。

1) 在 `internal/config/global.yaml` 中添加 embedding 模型配置：
```yaml
gemini:
  embedder: gemini-embedding-001
```

2) 设置 Google Gemini 的 API Key（`genai.NewClient` 会自动读取环境变量）：
```bash
export GOOGLE_API_KEY="<你的_API_KEY>"
```
> 若你使用的是其他兼容变量名，请参考 `google.golang.org/genai` 的文档；通常 `GOOGLE_API_KEY` 即可。

### 运行集成测试（Embedding）
该测试会真实调用 Gemini Embedding 接口，请确保已完成上述“配置”。

```bash
# 仅运行该用例，输出更详细日志
go test ./test -run TestGeminiEmbedder_Real -v
```

测试文件：`test/embedder_test.go`
- 读取 `internal/config/global.yaml` 的 `gemini.embedder` 作为嵌入模型（例如 `text-embedding-004`）
- 使用 `GOOGLE_API_KEY` 作为鉴权

### 目录结构（节选）
- `internal/config/viper.go`: `viper` 初始化与配置加载
- `internal/embadding/gemini/gemini.go`: Gemini Embedding 的封装
- `test/embedder_test.go`: 集成测试示例

### 常见问题
- 如果 `viper` 无法从环境变量解析到 `gemini.embedder`，请优先在 `internal/config/global.yaml` 中配置；当前实现主要面向文件配置。
- 若测试报鉴权错误，检查 `GOOGLE_API_KEY` 是否已正确导出到当前 shell。
