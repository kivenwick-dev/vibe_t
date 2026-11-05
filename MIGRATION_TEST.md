# 迁移测试分支：offline-report

基线：`6e95d7a4edc019e5f394d36a07feb69a52adf87e`（main）。此分支用于 GitHub → GitLab 迁移实验，暂不合并。

新增完全离线的资产报告命令和 report 模块，支持置信度、协议与 unknown 筛选，输出 Markdown 或 CSV。

## 后续增量修改

迁移开始后可继续修改本分支的实现、样例或本文档，再 commit/push。
记录迁移开始时间与后续提交 SHA，以便比较迁移目标的分支 tip、文件内容及提交历史。
本分支不会自动发起迁移、同步或合并。

## 手工验证

```bash
go test ./...
go vet ./...
go run ./cmd/report --input testdata/self-test.json --min-confidence 0.8 --format markdown
```
