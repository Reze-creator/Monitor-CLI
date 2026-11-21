# 服务健康探测 CLI

## 基本信息

- 学校：请填写
- 姓名：请填写
- 学号：请填写

## 已实现功能

- 从 `config.json` 读取 HTTP、HTTPS 和 TCP 探测目标。
- 并发执行目标探测，支持单次超时和最多 3 次失败重试。
- 支持 `--config`、`--timeout` 和 `-v` 命令行参数。
- 汇总成功率、响应时间分布和最慢目标，并生成 `monitor-log-时间戳.log`。

## 配置格式

```json
{
  "targets": [
    {
      "name": "示例网站",
      "type": "http",
      "address": "https://example.com",
      "expected_status": 200,
      "retry_count": 1
    },
    {
      "name": "本地数据库",
      "type": "tcp",
      "address": "localhost:3306"
    }
  ]
}
```

HTTP 目标可选 `expected_status`（不填时要求 2xx）、`contains`（检查响应正文是否包含指定文本）和 `user_agent`（为该目标设置 HTTP 请求的 `User-Agent`）。所有目标都可设置 `retry_count`，范围为 0 到 3。

## 运行

```text
go run .
go run . --config config.json --timeout 5 -v
go test ./...
```

## 开发中遇到的问题及处理

- 并发任务同时写入共享切片可能引发数据竞争。实现中每个任务把结果写入缓冲 channel，由主 goroutine 统一收集。
- 个别服务响应慢或无法连接会拖延整体执行。每次 HTTP/TCP 探测都设置超时，并且只按目标配置重试。
- 作业要求支持不同协议。通过目标的 `type` 字段选择 HTTP 或 TCP 探测逻辑，并在启动探测前校验配置。

每个并发任务开始探测前会等待 1 秒，这是作业要求的并发演示延迟。
