# MultiTools

MultiTools 是一个使用 Go 编写的多工具项目。当前第一阶段实现了
`url-visitor`：对已授权的一个或多个目标 URL 按固定全局速率发起持续访问，并输出基础可用性指标。

## 当前工具

### url-visitor

`url-visitor` 只应当用于你拥有、管理，或已经明确获得授权的 URL。工具默认采用保守策略：

- 只允许访问 `http` 和 `https` URL；
- 默认阻止内网、本机和私有网段目标；
- 如果配置了 `safety.allowed_hosts`，每个目标 host 都必须在白名单里；
- `rate_per_second` 不能超过 `safety.max_rate_per_second`；
- 默认启动前需要手动输入 `YES` 确认所有目标已授权。

## 运行方式

直接运行：

```bash
go run ./cmd/multitools url-visitor --config configs/urlvisitor.example.yaml
```

构建二进制后运行：

```bash
go build -o bin/multitools ./cmd/multitools
./bin/multitools url-visitor --config configs/urlvisitor.example.yaml
```

查看可用工具：

```bash
./bin/multitools --help
```

## url-visitor 配置说明

示例配置位于：[configs/urlvisitor.example.yaml](configs/urlvisitor.example.yaml)

```yaml
urls:
  - "https://example.com/"
  - "https://example.com/health"
strategy: "round_robin"
method: "GET"
rate_per_second: 10
concurrency: 10
duration: "1m"
timeout: "10s"
follow_redirects: true

headers:
  Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
  Accept-Language: "zh-CN,zh;q=0.9,en;q=0.8"

user_agents:
  - "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36"

safety:
  require_authorization_confirm: true
  max_rate_per_second: 10
  allowed_hosts:
    - "example.com"
  allow_private_networks: false
```

关键字段：

- `url`：要访问的单个目标 URL。为了兼容旧配置保留。
- `urls`：要访问的目标 URL 列表。配置多个目标时使用它。
- `strategy`：多个 URL 的分配策略。当前支持 `round_robin` 和 `random`。
- `round_robin`：轮询访问，每个 URL 尽量均匀。
- `random`：每次请求随机选择一个 URL。
- `method`：HTTP 方法，当前通常使用 `GET`。
- `rate_per_second`：全局每秒发起的请求数。这里的 `10` 表示所有 URL 合计每秒最多发起 10 次，不是每个 URL 10 次，也不是每个 worker 10 次。
- `concurrency`：并发 worker 数量，也就是最多同时处理多少个请求。
- `duration`：持续运行时间，例如 `1m`、`30s`。设置为 `0s` 表示一直运行，直到手动中断。
- `timeout`：单个请求的超时时间。
- `follow_redirects`：是否跟随 3xx 重定向。
- `headers`：附加请求头，会覆盖工具默认请求头中的同名字段。
- `user_agents`：User-Agent 列表，工具每次请求会从列表中随机选择一个。
- `safety.require_authorization_confirm`：启动前是否要求输入 `YES`。
- `safety.max_rate_per_second`：安全上限，防止误把访问速率调得过高。
- `safety.allowed_hosts`：目标 host 白名单，防止输错域名。
- `safety.allow_private_networks`：是否允许访问内网或本机地址。只有测试自己授权的内网服务时才应开启。

## 输出说明

运行时会每秒输出一次统计：

```text
[运行中] 目标数=2 总数=10 成功=10 HTTP失败=0 错误=0 下载=128.45 KB 平均每次下载=12.85 KB 状态码=200=10 平均耗时=74ms P95=110ms 最大耗时=132ms
  按目标统计：https://example.com/ 下载=80.12 KB 成功=5 HTTP失败=0 错误=0；https://example.com/health 下载=48.33 KB 成功=5 HTTP失败=0 错误=0
```

字段含义：

- `目标数`：本次运行配置的目标 URL 数量。
- `总数`：已完成请求总数。
- `成功`：HTTP 状态码小于 400 的响应数量。
- `HTTP失败`：HTTP 状态码大于等于 400 的响应数量，例如 `405`、`404`、`500`。
- `错误`：请求创建、发送、超时、DNS 等错误数量。
- `下载`：成功响应的内容总大小，会自动用 `B`、`KB` 或 `MB` 表示。
- `平均每次下载`：成功请求的平均响应体大小，适合判断单个图片或单个页面的平均大小。
- `按目标统计`：每个目标 URL 单独累计的下载大小、成功数和错误数。
- `状态码`：HTTP 状态码分布。
- `平均耗时`：平均响应耗时。
- `P95`：95 分位响应耗时。
- `最大耗时`：最大响应耗时。

下载大小统计的是响应体内容大小，不包含 HTTP 响应头，也不包含 TCP/TLS 协议开销。工具会请求 `Accept-Encoding: identity`，避免服务端返回压缩后的内容导致大小和你看到的原始图片大小不一致。如果目标站点因为反盗链、鉴权、重定向或 User-Agent 策略返回了较小的错误页，`按目标统计` 可以帮助定位具体哪个 URL 的返回内容异常。

## 项目结构

```text
cmd/multitools/              CLI 入口
internal/app/                应用分发层
internal/tools/              工具注册表和工具实现
internal/tools/urlvisitor/   url-visitor 工具实现
configs/                     示例配置文件
```

## 开发命令

格式化代码：

```bash
gofmt -w cmd internal
```

运行测试：

```bash
go test ./...
```

构建：

```bash
go build -o bin/multitools ./cmd/multitools
```
