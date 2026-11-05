# 迁移测试分支：client-formats

基线：`6e95d7a4edc019e5f394d36a07feb69a52adf87e`（main）。此分支用于 GitHub → GitLab 迁移实验，暂不合并。

客户端增加 CSV / JSONL 导入导出，保留 JSON 默认行为，并新增多格式编解码模块与样例。

## 后续增量修改

迁移开始后可继续修改本分支的实现、样例或本文档，再 commit/push。
记录迁移开始时间与后续提交 SHA，以便比较迁移目标的分支 tip、文件内容及提交历史。
本分支不会自动发起迁移、同步或合并。

## 手工验证

```bash
go test ./...
go vet ./...
go run ./cmd/client --input testdata/client.csv --input-format csv --output-format jsonl
```
