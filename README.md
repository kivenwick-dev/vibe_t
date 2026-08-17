# Banner Fingerprint

基于 Go 的批量 Banner 指纹识别系统，包含独立 server 与 client。

## 一键启动

```bash
docker compose up --build
```

Compose 会先等待 server 的 `/health` 通过真实健康检查，再启动 client 读取 `testdata/sample.json` 并输出识别结果。

## 本地运行

```bash
go run ./cmd/server
go run ./cmd/client --server http://localhost:8080 --input ./testdata/sample.json
```

规则文件默认位于 `rules/rules.json`，可通过 `RULES_PATH` 替换。规则包含协议、产品、版本和 OS hint 的正则提取逻辑，程序代码不需要随规则调整而修改。

## API

健康检查：

```bash
curl http://localhost:8080/health
```

批量识别：

```bash
curl -X POST http://localhost:8080/fingerprint \
  -H 'Content-Type: application/json' \
  --data @testdata/sample.json
```

无法识别的输入会返回 `protocol: "unknown"`、空产品和版本字段以及 `confidence: 0`，不会导致整个批次失败。
