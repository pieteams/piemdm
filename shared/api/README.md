本目录为 API 单一事实源（SSOT）
禁止提交生成代码
所有 client/server 代码必须由 scripts/gen-api-client.sh 生成

本仓库中仅“规范源文件”和“手动编写的库代码”允许使用 Git 跟踪。

以下内容【必须】提交到 Git：
- shared/api/swagger.yaml
- shared/api/swagger.json
- packages/go/openapi/** (这是手动编写的 OpenAPI 辅助库，包含 Auth/Middleware 等，非自动生成)
- packages/web/workflow-vue/** (这是手动编写的 工作流前端操作UI，非自动生成)
- 协议相关 README / 设计说明

以下内容【禁止】提交到 Git：
- packages/web/api-client/** (这是完全由 Swagger 生成的 TS 客户端)
- 任何由 OpenAPI / Proto / Schema 自动生成的代码

生成代码必须：
- 由 scripts/gen-api-client.sh 统一生成
- 可在任意环境中 100% 复现.