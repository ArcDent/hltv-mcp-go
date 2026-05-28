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
- 密钥推导: passphrase → SHA-256 → 32 字节 AES 密钥（与 person-summon `crypto.ts` 完全一致）
- passphrase 来源: `ENCRYPTION_KEY` 环境变量 或 `data/.encryption_key` 文件（存 hex 串，64 字符）
- IV: 每次加密随机生成 12 字节
- 存储格式: `base64(IV(12) + ciphertext + auth_tag(16))`
- 与 person-summon `crypto.ts` 格式兼容

```go
func Encrypt(plaintext string) (string, error)
func Decrypt(ciphertext string) (string, error)
```

### 2. 存储改造 `internal/http/handlers/translate.go`

- 配置文件路径改为 `./data/translate_config.json`（`data/` = `os.Getwd() + "/data"`，Docker 统一 `WORKDIR /`）
- 新增 `encrypted: true` 字段标记加密状态
- 加载时自动检测：明文旧格式自动加密回写升级
- 保存前自动加密 `api_key` 字段
- 启动时检查旧路径（`os.Executable()` 同级）是否存在旧格式 `translate_config.json`，存在则迁移到 `./data/` 并加密

### 3. 翻译代理 `POST /api/translate`

请求:
```json
{ "text": "...", "type": "title|article" }
```

响应:
```json
{ "translated": "..." }
```

内部流程:
1. 加载已保存的 provider_url / model / 解密 api_key
2. 检查是否已配置（configured），未配置返回 400
3. 根据 type 选择 system prompt:
   - `title`: `"将以下CS电竞新闻标题翻译为简体中文，只输出翻译结果，不要任何解释"`
   - `article`: `"将以下CS电竞新闻正文翻译为简体中文"`
4. 构造 OpenAI 兼容请求 → 调用 LLM API
5. 解析 `choices[0].message.content` → 返回译文

### 4. 前端简化

- `TranslateProvider.tsx`: 删除 `realKey`、`sessionStorage` 逻辑
- `NewsDetail.tsx`: `fetch('/api/translate', {body: {text, type:'article'}})` 替代直调 LLM
- `News.tsx`: `fetch('/api/translate', {body: {text, type:'title'}})` 替代直调 LLM

### 5. 密钥自动管理

启动时在 `main.go` 中执行，优先级:

1. `ENCRYPTION_KEY` 环境变量存在 → 作为 passphrase
2. `data/.encryption_key` 文件存在 → 读取其内容作为 passphrase
3. 都不存在 → 生成 64 字符随机 hex 串，写入 `data/.encryption_key` 作为 passphrase

passphrase 统一经过 SHA-256 推导 32 字节 AES 密钥，存入全局变量供 crypto 模块使用。

### 6. 部署变更

- Dockerfile: 创建 `/data` 目录（`RUN mkdir -p /data`），声明 `VOLUME ["/data"]`，添加 `WORKDIR /`
- docker-compose.yml: 挂载 `./data:/data` 持久化 volume
- `ENCRYPTION_KEY` 环境变量可选，不设则自动生成到 volume 中

## 文件变更清单

| 操作 | 文件 | 说明 |
|------|------|------|
| 新增 | `internal/crypto/crypto.go` | AES-256-GCM 加解密 |
| 新增 | `internal/crypto/crypto_test.go` | 加解密单元测试 |
| 修改 | `internal/http/handlers/translate.go` | 加密存储、路径迁移、翻译代理端点 |
| 修改 | `internal/http/router.go` | 注册 POST /api/translate |
| 修改 | `frontend/src/components/TranslateProvider.tsx` | 删除 realKey/sessionStorage |
| 修改 | `frontend/src/components/NewsDetail.tsx` | 走后端代理翻译 |
| 修改 | `frontend/src/pages/News.tsx` | 走后端代理翻译 |
| 修改 | `Dockerfile` | /data 目录 + WORKDIR / |
| 修改 | `docker-compose.yml` | data volume 挂载 |
| 修改 | `main.go` | 启动时密钥初始化（读 ENV/文件/自动生成） |
