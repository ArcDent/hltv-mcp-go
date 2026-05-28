# 翻译配置持久化与加密方案

## 问题

重启服务后翻译 key 无效化。根因是前端通过 `sessionStorage` 持有真实 API Key 直调 LLM API，重启/新会话后 sessionStorage 清空，但后端只返回脱敏 key，真实 key 无法恢复。

同时 `translate_config.json` 中 API Key 明文存储，存在安全风险。

## 方案

参照 person-summon 架构，将翻译请求从"前端直调 LLM"改为"后端代理"，API Key 加密存储，前端不接触真实 Key。

### 架构变更

```
之前: 浏览器 ─(直调,需真实key)─→ LLM API
之后: 浏览器 ─(POST /api/translate)─→ 后端 ─(解密key后转发)─→ LLM API
```

## 模块设计

### 1. 加密模块 `internal/crypto/crypto.go`

- 算法: AES-256-GCM（Go 标准库 `crypto/aes` + `crypto/cipher`）
- Key: `ENCRYPTION_KEY` 环境变量 → SHA-256 → 32 字节
- IV: 每次加密随机生成 12 字节
- 格式: `base64(IV + ciphertext + tag)`
- 与 person-summon `crypto.ts` 格式兼容

```go
func Encrypt(plaintext string) (string, error)
func Decrypt(ciphertext string) (string, error)
```

### 2. 存储改造 `internal/http/handlers/translate.go`

- 配置文件路径改为 `./data/translate_config.json`（统一持久化目录）
- 新增 `encrypted: true` 字段标记加密状态
- 加载时自动检测：明文旧格式自动加密回写升级
- 保存前自动加密 `api_key` 字段

### 3. 翻译代理 `POST /api/translate`

请求:
```json
{ "text": "...", "type": "title|article" }
```

响应:
```json
{ "translated": "..." }
```

内部流程: 加载配置 → 解密 key → 构造 OpenAI 兼容请求 → 调用 LLM → 返回译文

### 4. 前端简化

- `TranslateProvider.tsx`: 删除 `realKey`、`sessionStorage` 逻辑
- `NewsDetail.tsx`: `fetch('/api/translate', ...)` 替代直调 LLM
- `News.tsx`: 同上

### 5. 密钥自动管理

启动优先级:
1. `ENCRYPTION_KEY` 环境变量
2. `data/.encryption_key` 文件
3. 自动生成随机 32 字节写入 `data/.encryption_key`

### 6. 部署变更

- Dockerfile: 创建 `/data` 目录，声明 VOLUME
- docker-compose.yml: 挂载 `./data:/data` 持久化 volume
- `ENCRYPTION_KEY` 可选，不设则自动生成

## 文件变更清单

| 操作 | 文件 |
|------|------|
| 新增 | `internal/crypto/crypto.go` |
| 新增 | `internal/crypto/crypto_test.go` |
| 修改 | `internal/http/handlers/translate.go` |
| 修改 | `internal/http/router.go` |
| 修改 | `frontend/src/components/TranslateProvider.tsx` |
| 修改 | `frontend/src/components/NewsDetail.tsx` |
| 修改 | `frontend/src/pages/News.tsx` |
| 修改 | `Dockerfile` |
| 修改 | `docker-compose.yml` |
| 修改 | `main.go`（加密密钥初始化） |
