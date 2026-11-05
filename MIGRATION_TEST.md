# 迁移测试分支：rules-catalog

基线：`6e95d7a4edc019e5f394d36a07feb69a52adf87e`（main）。此分支用于 GitHub → GitLab 迁移实验，暂不合并。

新增 SMTP、IMAP、POP3、Memcached、MQTT、AMQP、Caddy 和 lighttpd 规则及样例；新增规则校验与目录命令。

## 后续增量修改

迁移开始后可继续修改本分支的实现、样例或本文档，再 commit/push。
记录迁移开始时间与后续提交 SHA，以便比较迁移目标的分支 tip、文件内容及提交历史。
本分支不会自动发起迁移、同步或合并。

## 手工验证

```bash
go test ./...
go vet ./...
go run ./cmd/rulecheck --rules rules/rules.json
```
