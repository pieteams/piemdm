# SDK 下載與參考

為了方便開發人員快速集成 PieMDM OpenAPI，我們提供了多種主流編程語言的 SDK 封裝。這些 SDK 已經處理了複雜的簽名生成、請求構建及錯誤處理邏輯。

## 1. SDK 下載

您可以直接下載以下語言的 SDK 源文件並集成到您的項目中：

| 編程語言 | 下載鏈接 | 說明 |
| :--- | :--- | :--- |
| **Go** | [piemdm_openapi.go](/sdk/go/piemdm_openapi.go) | 基於 `net/http` 標準庫實現 |
| **Java** | [PiemdmOpenApi.java](/sdk/java/PiemdmOpenApi.java) | 單文件實現，依賴標準 HTTP 庫 |
| **PHP** | [PiemdmOpenApi.php](/sdk/php/PiemdmOpenApi.php) | 原生 PHP 實現，兼容 7.x/8.x |
| **JavaScript** | [piemdm-openapi.js](/sdk/javascript/piemdm-openapi.js) | 適用於 Node.js 和瀏覽器環境 |

> [!TIP]
> **關於 Go 獲取方式**：
> 如果您在 PieMDM 相關的 Monorepo 中進行開發，推薦使用 Go Modules 方式直接引用：
> `go get github.com/pieworks/piemdm/packages/go/openapi`

---

## 2. 使用前準備

在使用 SDK 之前，請確保您已獲得以下必要信息：

1. **AppID**: 系統分配的應用標識。
2. **AppSecret**: 應用密鑰（**請勿洩露**）。
3. **BaseURL**: OpenAPI 接口的基礎訪問地址（例如 `https://piemdm.example.com`）。

您需要確保發起請求的服務器 IP 已被列入 PieMDM 後台應用的 **IP 白名單**中。

---

## 3. 核心功能

SDK 封裝了以下核心能力，確保請求符合安全規範：

- **自動化簽名 (Self-Signing)**: 根據 `AppSecret` 自動計算每筆請求的 HMAC-SHA256 簽名。
- **防止重放 (Anti-Replay)**: 自動生成 `Timestamp` 和唯一 `Nonce`。
- **標準化響應**: 統一解析 API 響應格式，簡化異常捕獲處理。

有關各個 API 的具體參數說明，請參考 [API 接口文檔](./open-api)。

---

## 4. 示例代碼

詳細的調用示例（包括 SDK 調用和手動簽名邏輯）請查看：
👉 **[OpenAPI 調用示例](./call-example)**
