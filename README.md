# WorkBuddy2API Advance

<p align="center">
  <b>把腾讯 CodeBuddy 账号变成 OpenAI 兼容 API 的多账号网关 · 附 Web 管理面板</b><br>
  Web 面板 · OAuth 浏览器登录 · 账号池轮转 · 熔断与冷却 · 会话粘性 · 定时签到 / 活跃 / 旅行 / 保活<br>
  <b>客户端密钥体系：随时增删 · 有效期（小时）· 模型白名单</b> · 流式 / 非流式
</p>

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.22.5-00ADD8?logo=go&logoColor=white&style=flat-square">
  <img alt="API" src="https://img.shields.io/badge/API-OpenAI_Compatible-412991?style=flat-square">
  <img alt="Deploy" src="https://img.shields.io/badge/Deploy-Single_Binary%20%7C%20Docker-2496ED&style=flat-square">
  <img alt="Keys" src="https://img.shields.io/badge/Client_Keys-%E5%A2%9E%E5%88%A0%C2%B7%E6%9C%89%E6%95%88%E6%9C%9F%C2%B7%E7%99%BD%E5%90%8D%E5%8D%95-4A90D9&style=flat-square">
  <img alt="Transport" src="https://img.shields.io/badge/Transport-SSE%20%2F%20Streaming-0DBD8B&style=flat-square">
</p>

---

> **本项目是 [Sliverkiss/workbuddy2api](https://github.com/Sliverkiss/workbuddy2api) 的增强分支**（fork）。
> 上游仓库现已删除；本项目已同步至上游删库前的最后一次更新（`ea8b1e5`），此后由本分支独立维护演进。
> 本分支（Advance）在上游与运维层重构的基础上，新增**客户端密钥体系**：任意新建 / 删除 API Key，每把密钥可独立设置有效期与可用模型白名单，适合把网关分享给多个设备 / 多个客户端使用而无需共用主密钥。

> ⚠️ **本项目为二次开源开发，在原项目基础上增加了 API Key 的搭建与管理；所创建的 API Key 仅支持个人使用。本作者强烈谴责对外提供付费 API Key 的行为，二次开发者对本项目的任何修改皆与本作者无关。** 详见 [使用声明](#使用声明必读)。

## 使用声明（必读）

本项目（含上游 Sliverkiss/workbuddy2api）的开发初衷只有一个：**方便个人管理自己的 CodeBuddy 账号**——自动签到、保活、给自己的本地工具提供一个 OpenAI 兼容入口。它是免费、开源、按「原样」提供的个人自用工具。

1. **一切商用 / 售卖行为与本项目及作者无关。** 本项目不授权、不支持、不参与任何面向公众的 API 售卖、账号池出租、卡密收费分发；行为人由此产生的一切后果（包括但不限于账号封禁、条款违约与法律风险）由其自行承担。
2. **批量注册与转售接口配额违反目标平台服务条款。** 上游仓库已删除、停止公开维护，请勿再消耗社区的善意。
3. **请勿购买任何「收费版」「卡密版」「公益中转版」。** 本项目永远免费开源。任何加壳、加密、捆绑收费的「版本」都是他人篡改的产物；此类分发无法审计，存在被植入后门、窃取你 CodeBuddy 凭证的风险（`auths/` 保存的是明文 accessToken / refreshToken）。
4. **关于开源协议**：本项目基于 MIT 协议开源，协议赋予代码层面的自由，但不赋予以本项目名义宣传、售卖、捆绑分发的权利。作者不为任何第三方分发版本提供支持。
5. **作者保留止损的权利。** 若滥用行为持续，作者可能随时停止维护、关闭或删除仓库，且不另行通知。

如果你的用途是管理自己的账号，欢迎正常使用、反馈问题与提交 PR。

## 项目简介

WorkBuddy2API 是一个自托管的 **OpenAI 兼容反向代理网关**，将腾讯 CodeBuddy（`copilot.tencent.com`）账号包装为统一的 `/v1/chat/completions` 服务：

- 官方不提供 OpenAI 形态的开放 API，本项目通过 **OAuth 设备授权**（面板「添加账号」或 `login.sh`）获取账号凭证，在网关侧做 token 自动刷新、账号池调度与流量治理；
- 面向**个人多账号**场景：多账号共享、单号故障自动换号、冷却 / 熔断防止雪崩、会话粘性保证多轮上下文不跳号；
- 对客户端只暴露 OpenAI 兼容接口，现有 SDK / 前端 / 工具**零改造接入**；
- **本分支新增客户端密钥体系**：一把主密钥（管面板）+ 若干客户端密钥（调模型），每把客户端密钥独立可控、可过期、可限制模型。

## 核心能力

| 能力 | 说明 |
|---|---|
| 🔑 **OAuth 一键登录** | `login.sh` 设备授权流程，或面板「添加账号」浏览器内完成，凭证落盘并热加载 |
| 🔄 **多账号池** | 快过期积分加权 + 成本分层 + 加权随机选号，Top-5 候选 + 防惊群 |
| 🛡️ **熔断与冷却** | 429 软冷却 600s 起指数退避、404 固定 60s 短冷却、402 硬冷却至次日 04:00、连续失败熔断、在途租约限流 |
| 🧲 **会话粘性** | 同一会话尽量绑定同一账号，TTL 滚动续期，失败自动解绑 |
| ⏰ **定时任务** | 签到（09/21 点，含连登管家）+ 活跃上报（10 点）+ 猫猫旅行（09/21 点）+ token 保活（22 点），四类独立开关 |
| ⚡ **流式 + 非流式** | 出站强制 `stream:true`；SSE 帧按规范白名单重建；非流式由本地聚合为单响应 |
| 🧠 **推理模型兼容** | 思维链注入（`thinking.type=enabled`）、`reasoning_content` 多轮回填、effort 档位自动降级 |
| 💬 **系统提示词体系** | 网关自有提示词替换客户端 system（默认 `custom`），从源头消灭 system 来源的内容误报 |
| 🗑️ **指纹脱敏** | 出站请求体黑名单指纹字段清洗（可关闭） |
| 📊 **可观测** | 每请求一行表格日志（TTFB / token 速率 / uid）+ 调用来源 IP·UA；`/healthz` 带 `service` 身份标识 |
| 💾 **状态持久化** | 池状态本地原子落盘 + Upstash Redis 异步镜像（可选），重启择新恢复 |
| 🖥️ **Web 管理面板** | 内嵌单页面板（明暗主题）：账号运维 / 用量与积分分析 / 模型档位查询 / 在线改配置（热生效）/ 运行日志 / 积分任务 |
| 🗝️ **客户端密钥体系**（本分支新增） | 任意新建 / 删除 API Key；每把密钥可设有效期（小时，到点自动失效）与可用模型白名单，详见下文 |

## 🗝️ 客户端密钥体系（本分支新增）

主 `api_key` 只用于登录面板。日常客户端调模型，一律使用面板里新建的**客户端密钥**——每把密钥有三个可控维度：

| 维度 | 行为 |
|---|---|
| **随时增删** | 面板一键新建 / 删除，热生效、无需重启；密钥落盘 `config.json` 的 `client_keys`，重启不丢 |
| **有效期（小时）** | 新建时填 `N` 小时，到点自动失效；留空或 `0` = 永久；上限 8760 小时（1 年） |
| **模型白名单** | 新建时从上游实时模型列表勾选可用模型；不勾 = 不限制。越权调用返回 `403 model_not_allowed`，且 `/v1/models` 只返回白名单内模型 |

设计要点：

- **权限隔离**：客户端密钥只能调 `/v1/*` 模型接口，**不能登录面板**（面板只用主 `api_key`）；
- **删除即时生效**：删除后下一个请求立刻 `401 invalid_api_key`，无需重启、无后台清理任务；
- **过期即时生效**：到点后下一个请求 `401`，与删除同样处理；
- **越权零消耗**：模型白名单在动用账号池**之前**拦截，越权请求不消耗任何账号额度；
- **模型名归一**：白名单里勾 `cn:glm-5.2`，客户端填裸名 `glm-5.2` 也放行；`global:` 前缀模型（如 `global:gpt-6-luna`）只在 global 域匹配。

### 面板操作

面板 → **配置** → 「客户端密钥」卡片：

1. 填名称（如「笔记本 / Roo Code」）；
2. 「有效期(小时)，0=永久」按需填写（默认留空 = 永久）；
3. 点「**选择可用模型**」展开选择器——实时探测上游，列出账号池当前全部可用模型（与 `/v1/models` 同口径），勾选允许的模型（有全选 / 清空）；
4. 「新建密钥」——自动生成 `sk-` 开头的密钥并复制到剪贴板（**原文只显示这一次**，之后表格只显示掩码）。

表格列：名称 / 密钥掩码 / 创建时间 / 到期（永久、剩余时间如「剩 59分」、或红色「已过期」）/ 可用模型（全部或前几个 id）。随时可复制、删除；过期密钥仍显示在列表中，手动删除即可清理。

### 脚本化调用

面板接口与页面同 Bearer 鉴权（用主 `api_key`）：

```bash
# 新建：24 小时有效、限用 glm-5.2 与 deepseek-v4-pro
curl -X POST http://127.0.0.1:7863/panel/api/client_keys \
  -H "Authorization: Bearer 主api_key" \
  -H "Content-Type: application/json" \
  -d '{"name":"笔记本","expires_in_hours":24,"models":["cn:glm-5.2","cn:deepseek-v4-pro"]}'

# 列表
curl http://127.0.0.1:7863/panel/api/client_keys -H "Authorization: Bearer 主api_key"

# 删除
curl -X POST http://127.0.0.1:7863/panel/api/client_keys/delete \
  -H "Authorization: Bearer 主api_key" \
  -d '{"key":"sk-要删的密钥"}'
```

`config.json` 中的字段形态：

```json
{
  "api_key": "主密钥，管面板用",
  "client_keys": [
    {
      "name": "笔记本",
      "key": "sk-...",
      "created_at": 1729100000,
      "expires_at": 1729186400,
      "models": ["cn:glm-5.2", "cn:deepseek-v4-pro"]
    }
  ]
}
```

> **前提**：`api_key` 必须非空。为空表示网关整体不鉴权，客户端密钥不生效（面板会提示）。

## 快速开始

### 方式一：Docker Compose（推荐服务器部署）

```bash
# 1. 克隆
git clone https://github.com/zheyudu39-del/workbuddy2api-advance.git
cd workbuddy2api-advance

# 2. 准备配置（compose 挂载此文件，缺失会导致容器启动失败）
cp config.example.json config.json   # 建议编辑 config.json 设置 api_key

# 3. 启动（首次会构建镜像，约 1-2 分钟）
docker compose up -d

# 4. 健康检查（无可用账号时返回 503）
curl -s http://127.0.0.1:7863/healthz
# {"healthy":0,"total":0,"service":"workbuddy2api"}
```

### 方式二：Windows 单文件运行（无需 Docker）

下载 Release 中的 `wb2api.exe`，或从源码构建（`go build ./cmd/server`）：

```powershell
# 直接运行：首次启动自动生成 config.json（含随机 api_key，日志打印一次）
.\wb2api.exe

# 浏览器打开面板添加账号
# http://127.0.0.1:7863/panel/
```

### 添加账号（登录）

```bash
./login.sh
# 按提示在浏览器打开授权链接 → 回到终端确认 → 凭证落盘 auths/workbuddy-<uid>.json
```

或在面板点「添加账号」，浏览器内完成 OAuth 设备授权，登录后**热加载进池，免重启**。

### 验证

```bash
# 模型列表（用主 api_key 或任一客户端密钥）
curl http://127.0.0.1:7863/v1/models -H "Authorization: Bearer sk-..."

# 流式聊天
curl -N http://127.0.0.1:7863/v1/chat/completions \
  -H "Authorization: Bearer sk-..." -H "Content-Type: application/json" \
  -d '{"model":"cn:glm-5.2","messages":[{"role":"user","content":"hi"}]}'

# 账号状态（汇总 + 每账号详情）
curl http://127.0.0.1:7863/status -H "Authorization: Bearer 主api_key"
```

## 配置说明

`config.example.json` 是配置项最完整的参考（示例值一律为占位符，不含真实密钥）。关键字段：

| 字段 | 默认 | 说明 |
|---|---|---|
| `listen` | `:7863` | HTTP 监听地址 |
| `api_key` | 空 | 网关 + 面板主密钥；**空 = 不鉴权直接放行**（公网必须设置） |
| `client_keys` | `[]` | 客户端密钥列表（name/key/created_at/expires_at/models），面板维护，热生效 |
| `auth_dir` | `./auths` | 账号凭证目录 |
| `state_file` | `./data/state.json` | 账号池状态持久化文件 |
| `schedule.checkin_hours` | `[9, 21]` | 每日本地时区整点签到 + 余额查询解冻 |
| `schedule.travel_hours` | `[9, 21]` | 猫猫旅行状态机（领养 / 派出 / 领奖） |
| `schedule.activity_hours` | `[10]` | 对话活跃上报 |
| `schedule.keepalive_hours` | `[22]` | token 保活 |
| `schedule.*_enabled` | `true` | 四类任务总开关 |
| `upstream.timeout_seconds` | `120` | 短 RPC（刷新 / 签到 / 余额 / 模型列表）总时长上限 |
| `upstream.header_timeout_seconds` | 回落 `timeout_seconds` | 聊天首字节前上限 |
| `upstream.idle_timeout_seconds` | `300` | 聊天流中空闲上限（活跃续命，静默断流） |
| `prompt.mode` | `custom` | 系统提示词模式：`custom` / `append` / `passthrough` |
| `pool.*` | 见样例 | 熔断 / 在途 / 权重 / 积分保底等流量治理参数 |
| `session_sticky.*` | 见样例 | 会话粘性路由（开关 / TTL / GC 周期） |
| `logging.*` | 见样例 | 请求归档开关 / 保留天数 / 容量上限 / 来源 IP·UA 记录 |

环境变量覆盖（`WB2A_*`，非空才覆盖）：`WB2A_LISTEN`、`WB2A_API_KEY`、`WB2A_AUTH_DIR`、`WB2A_STATE_FILE`、`WB2A_USER_AGENT`、`WB2A_PROMPT_MODE` 等。

**配置热生效**：面板保存配置后，`api_key`、`client_keys`、`cooldown.*`、`pool.*`、`schedule.*`、`logging.*` 立即生效；`listen`、`auth_dir`、`state_file`、`upstream.*`、`upstash.*` 需重启。

## API 端点

| 端点 | 鉴权 | 说明 |
|---|---|---|
| `POST /v1/chat/completions` | Bearer（主密钥或任一有效客户端密钥） | OpenAI 兼容补全；流式 / 非流式；请求体上限 8 MiB；模型越权返回 403 |
| `GET /v1/models` | Bearer | 模型列表（纯动态拉取，缓存 1h）；**客户端密钥只返回其白名单内模型** |
| `GET /status` | Bearer | 账号状态汇总 + 每账号详情 + `model_locks` 限流视图 |
| `GET /healthz` | 无 | 健康检查：返回 `{"healthy":N,"total":M,"service":"workbuddy2api"}` |
| `GET /panel/api/*` | Bearer（仅主 api_key） | 面板数据接口（overview / usage / models / logs / config / client_keys 等） |

模型名规则：不带前缀按 **CN 域**解析（`glm-5.2` ≡ `cn:glm-5.2`）；`global:` 前缀走国际域（如 `global:gpt-6-luna`）。两域模型名单不同，可先用 `/v1/models` 查可用 id。

## Web 管理面板

```
http://127.0.0.1:7863/panel/
```

鉴权与 API 同口径：`api_key` 非空时面板要求输入一次主密钥（浏览器 localStorage 记住）；为空则直接可用。明暗主题切换，左侧导航：

| 视图 | 功能 |
|---|---|
| **账号池** | 统计条 + 账号表（健康色条 / 状态 / 积分量条 / 成功失败 / 在途 / 最近用量）；单号与批量操作（签到 / 余额 / 任务 / 解冻 / 禁用 / 移除）；模型锁池表 |
| **用量** | 指标卡 + Token 时序图 + 用量明细（按账号 / 模型 / 域）+ 积分扣除历史 |
| **积分构成** | 积分到期分布 + 账号对比 + 逐包明细 |
| **任务中心** | 成长任务队列：一键扫描全部账号待办，按账号排队执行，进度实时更新 |
| **模型与档位** | 实时查询上游：每模型积分倍率、思考档位、上下文长度与最大输出；支持按条件查询与排序 |
| **配置** | 在线编辑 config.json（服务 / 定时任务 / 账号池 / 上游 / 日志）+ **客户端密钥管理** |
| **运行日志** | 请求记录表（时间 / 结果 / 模型 / 账号 / 来源 IP / UA / 耗时 / Token / 积分），支持筛选与时间范围 |

顶部有「刷新」（全量刷新真实余额）与「添加账号」（浏览器内 OAuth）。

## Claude Code（Anthropic 协议）接入

Claude Code 使用 Anthropic 协议（`/v1/messages`），网关原生只讲 OpenAI 协议，中间需要一座协议桥。仓库未内置桥脚本，可自备一个 Anthropic↔OpenAI 转发器（监听本机、把 `/v1/messages` 翻译为 `/v1/chat/completions` 打到网关 `UPSTREAM`），然后在 Claude Code 侧配置：

| 字段 | 值 |
|---|---|
| `ANTHROPIC_BASE_URL` | 桥的地址（如 `http://127.0.0.1:8080`） |
| `ANTHROPIC_AUTH_TOKEN` | 任意非空（桥层不校验；真正的鉴权在桥转发时带上网关密钥） |
| 模型 | 任意 claude 名字（由桥映射到上游真实模型，如 `cn:glm-5.2`） |

OpenAI 入口不受任何影响——ZCode / Cherry Studio / Roo Code / SDK 继续走 `/v1/chat/completions`，两条线共用账号池与额度。

## 部署运维

### 安全与合规

1. **凭据管理**：`auths/` 保存明文 accessToken / refreshToken，`config.json` 含 api_key——**两者均已纳入 `.gitignore`，切勿提交**；
2. **网络暴露**：服务自身只提供明文 HTTP，公网部署**务必置于 HTTPS 反代之后**（Nginx/Caddy）并限流。反代需关缓冲（`proxy_buffering off`）并放宽超时（`proxy_read_timeout 3600s`），否则流式输出会「憋很久一次性蹦出」；
3. **日志敏感度**：请求归档不记录提示词、响应正文与 Authorization；来源 IP·UA 由 `logging.request_client_info` 控制（默认开，可关）；
4. **鉴权实现**：`internal/httpauth` 统一 server 与 panel 的 Bearer 校验，SHA-256 摘要 + 常量时间比较。

### 常见问题

- **面板提示「api_key 为空（未启用鉴权）」**：客户端密钥不生效，先在 `config.json` 设主 `api_key`；
- **调用返回 503**：账号池无可用号或该域没有此模型——模型名前加 `global:` 再试；
- **调用返回 403 model_not_allowed**：该密钥白名单不含此模型，改用白名单内模型或重建密钥；
- **流式在反代后卡顿**：反代未关 `proxy_buffering`；
- **Docker 登录后写入 auths 报 permission denied**：按 `.env` 设 `PUID`/`PGID` 为挂载目录属主，或给挂载目录放行。

## 免责声明

本项目为非官方网关，仅限本人授权账号、本机 / 私有环境测试。使用本项目产生的一切后果由使用者自行承担，与作者及贡献者无关。

## License

MIT
