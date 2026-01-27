# SDK Download & Reference

To facilitate developers in quickly integrating PieMDM OpenAPI, we provide SDK encapsulations for multiple mainstream programming languages. These SDKs handle complex signature generation, request construction, and error handling logic.

## 1. SDK Download

You can directly download the SDK source files for the following languages and integrate them into your project:

| Language | Download Link | Description |
| :--- | :--- | :--- |
| **Go** | [piemdm_openapi.go](/sdk/go/piemdm_openapi.go) | Based on `net/http` standard library |
| **Java** | [PiemdmOpenApi.java](/sdk/java/PiemdmOpenApi.java) | Single file implementation, depends on standard HTTP library |
| **PHP** | [PiemdmOpenApi.php](/sdk/php/PiemdmOpenApi.php) | Native PHP implementation, compatible with 7.x/8.x |
| **JavaScript** | [piemdm-openapi.js](/sdk/javascript/piemdm-openapi.js) | Suitable for Node.js and browser environments |

> [!TIP]
> **About Go Retrieval**:
> If you are developing within the PieMDM related Monorepo, it is recommended to reference directly using Go Modules:
> `go get github.com/pieworks/piemdm/packages/go/openapi`

---

## 2. Preparation Before Use

Before using the SDK, please ensure you have obtained the following necessary information:

1. **AppID**: System-assigned application identifier.
2. **AppSecret**: Application secret key (**Do not leak**).
3. **BaseURL**: The base access address of the OpenAPI interface (e.g., `https://piemdm.example.com`).

You need to ensure that the server IP initiating the request has been added to the **IP Whitelist** of the application in the PieMDM backend.

---

## 3. Core Features

The SDK encapsulates the following core capabilities to ensure requests meet security standards:

- **Self-Signing**: Automatically accepts the `AppSecret` to calculate the HMAC-SHA256 signature for each request.
- **Anti-Replay**: Automatically generates `Timestamp` and unique `Nonce`.
- **Standardized Response**: Unifies API response format parsing, simplifying exception handling.

For specific parameter descriptions of each API, please refer to the [API Documentation](./open-api).

---

## 4. Example Code

For detailed calling examples (including SDK calls and manual signature logic), please view:
👉 **[OpenAPI Call Examples](./call-example)**
