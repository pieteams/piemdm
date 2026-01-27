# SDK 下载与参考

为了方便开发人员快速集成 PieMDM OpenAPI，我们提供了多种主流编程语言的 SDK 封装。这些 SDK 已经处理了复杂的签名生成、请求构建及错误处理逻辑。

## 1. SDK 下载

您可以直接下载以下语言的 SDK 源文件并集成到您的项目中：

| 编程语言 | 下载链接 | 说明 |
| :--- | :--- | :--- |
| **Go** | [piemdm_openapi.go](/sdk/go/piemdm_openapi.go) | 基于 `net/http` 标准库实现 |
| **Java** | [PiemdmOpenApi.java](/sdk/java/PiemdmOpenApi.java) | 单文件实现，依赖标准 HTTP 库 |
| **PHP** | [PiemdmOpenApi.php](/sdk/php/PiemdmOpenApi.php) | 原生 PHP 实现，兼容 7.x/8.x |
| **JavaScript** | [piemdm-openapi.js](/sdk/javascript/piemdm-openapi.js) | 适用于 Node.js 和浏览器环境 |

> [!TIP]
> **关于 Go 获取方式**：
> 如果您在 PieMDM 相关的 Monorepo 中进行开发，推荐使用 Go Modules 方式直接引用：
> `go get github.com/pieworks/piemdm/packages/go/openapi`

---

## 2. 使用前准备

在使用 SDK 之前，请确保您已获得以下必要信息：

1. **AppID**: 系统分配的应用标识。
2. **AppSecret**: 应用密钥（**请勿泄露**）。
3. **BaseURL**: OpenAPI 接口的基础访问地址（例如 `https://piemdm.example.com`）。

您需要确保发起请求的服务器 IP 已被列入 PieMDM 后台应用的 **IP 白名单**中。

---

## 3. 核心功能

SDK 封装了以下核心能力，确保请求符合安全规范：

- **自动化签名 (Self-Signing)**: 根据 `AppSecret` 自动计算每笔请求的 HMAC-SHA256 签名。
- **防止重放 (Anti-Replay)**: 自动生成 `Timestamp` 和唯一 `Nonce`。
- **标准化响应**: 统一解析 API 响应格式，简化异常捕获处理。

有关各个 API 的具体参数说明，请参考 [API 接口文档](./open-api)。

---

## 4. 示例代码

详细的调用示例（包括 SDK 调用和手动签名逻辑）请查看：
👉 **[OpenAPI 调用示例](./call-example)**
