# Banner Fingerprint

基于 Go 的批量 Banner 指纹识别系统，包含独立 server 与 client。

## 一键启动

```bash
docker compose up --build
```

Compose 会先等待 server 的 `/health` 通过真实健康检查，再启动 client 读取 `testdata/sample.json` 并输出识别结果。
若本机 8080 已被占用，可指定另一端口：

```bash
HOST_PORT=18080 docker compose up --build
```

## 本地运行

```bash
go run ./cmd/server
go run ./cmd/client --server http://localhost:8080 --input ./testdata/sample.json
```

规则文件默认位于 `rules/rules.json`，可通过 `RULES_PATH` 替换。规则包含协议、产品、版本和 OS hint 的正则提取逻辑，程序代码不需要随规则调整而修改。

使用镜像处理自定义本地文件：

```bash
docker compose run --rm -v "$PWD/input.json:/data/input.json:ro" client \
  --server http://server:8080 --input /data/input.json
```

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

请求体是 JSON 数组：

```json
[
  {"ip":"1.2.3.4","port":22,"banner":"SSH-2.0-OpenSSH_8.9p1 Ubuntu"}
]
```

响应中的每一项始终包含 `ip`、`port`、`protocol`、`product`、`version`、`os_hint`、`confidence`。无法识别时返回：

```json
{"ip":"1.2.3.99","port":9999,"protocol":"unknown","product":"","version":"","os_hint":"","confidence":0}
```

无法识别的输入会返回 `protocol: "unknown"`、空产品和版本字段以及 `confidence: 0`，不会导致整个批次失败。

## Migration test: api-summary

新增批量汇总 API，统计协议、平均置信度、唯一主机与未知结果，并校验输入和尾随 JSON。

```bash
curl -X POST http://localhost:8080/fingerprint/summary -H 'Content-Type: application/json' --data @testdata/sample.json
```
