# 响应处理函数 (resp)

## 📋 目录概述

`api/pkg/helper/resp` 目录包含通用的 HTTP 响应处理函数，主要负责控制 API 响应的结构和格式。

## 🎯 核心职责

### 1. **响应结构控制**
- 统一 API 响应的数据格式
- 设置标准的 HTTP 响应头部
- 提供成功和错误的响应处理函数

### 2. **HTTP 头部管理**
- 设置 Content-Type、Accept 等标准头部
- 配置跨域（CORS）相关头部
- 添加速率限制（Rate Limit）头部信息
- 生成请求 ID（X-Request-Id）用于追踪

### 3. **分页支持**
- 生成分页链接（Link 头部）
- 支持标准的 GitHub API 风格分页

## 🔧 主要函数

### `HandleSuccess(c *gin.Context, data interface{})`
处理成功的 HTTP 响应。

**参数**：
- `c *gin.Context`: Gin 上下文
- `data interface{}`: 响应数据

**功能**：
- 设置标准 HTTP 头部
- 返回 HTTP 200 状态码
- 包装响应数据

### `HandleError(c *gin.Context, httpCode int, message string, data interface{})`
处理错误的 HTTP 响应。

**参数**：
- `c *gin.Context`: Gin 上下文
- `httpCode int`: HTTP 状态码
- `message string`: 错误消息
- `data interface{}`: 附加错误数据

**功能**：
- 设置标准 HTTP 头部
- 返回指定的 HTTP 状态码
- 提供结构化的错误信息

### `SetHeader(c *gin.Context)`
设置标准的 HTTP 响应头部。

**功能**：
- Content-Type: application/json; charset=utf-8
- CORS 相关头部
- 速率限制头部
- 请求追踪 ID

### `GeneratePaginationLinks(r *http.Request, page, pageSize, total int) linkheader.Links`
生成分页链接。

**参数**：
- `r *http.Request`: HTTP 请求
- `page int`: 当前页码
- `pageSize int`: 每页大小
- `total int`: 总记录数

**返回**：
- `linkheader.Links`: 分页链接集合

## 📝 使用示例

### 基本使用
```go
import "piemdm/pkg/helper/resp"

// 成功响应
func GetUser(c *gin.Context) {
    user := &User{ID: 1, Name: "张三"}
    resp.HandleSuccess(c, user)
}

// 错误响应
func CreateUser(c *gin.Context) {
    err := validateUser(c)
    if err != nil {
        resp.HandleError(c, http.StatusBadRequest, "用户数据验证失败", nil)
        return
    }
    // ... 创建用户逻辑
}
```

### 分页使用
```go
import "piemdm/pkg/helper/resp"

func ListUsers(c *gin.Context) {
    page := 1
    pageSize := 20
    total := 100

    // 生成分页链接
    links := resp.GeneratePaginationLinks(c.Request, page, pageSize, total)

    // 设置 Link 头部
    c.Header("Link", links.String())

    // 返回数据
    users := []User{/* ... */}
    resp.HandleSuccess(c, users)
}
```

## 🔄 与 response 目录的关系

### 分工协作
- **`resp`（本目录）**: 控制响应的**结构和格式**
- **`internal/transport/response`**: 定义响应的**数据结构**

- `resp` 包：**通用工具**，负责**格式化**和**标准化** API 响应
- `response` 包：**数据定义**，负责定义 API 返回的具体**数据字段**

### 示例

在 Controller 中使用 `piemdm/pkg/helper/resp` 来包装 `piemdm/internal/transport/response` 定义的数据结构：

```go
import (
    "piemdm/pkg/helper/resp"
    "piemdm/internal/transport/response"
)

func GetUserDetail(c *gin.Context) {
    // 获取业务数据
    user := getUserFromDB()

    // 转换为响应结构体
    userResp := response.UserResponse{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
        // ... 其他字段
    }

    // 使用 resp 函数返回
    resp.HandleSuccess(c, userResp)
}
```

## ⚙️ 配置说明

### HTTP 头部配置
默认配置的 HTTP 头部包括：

1. **CORS 头部**：
   - `Access-Control-Allow-Origin: *`
   - `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
   - `Access-Control-Allow-Headers: Content-Type, Authorization`

2. **速率限制头部**：
   - `X-Ratelimit-Limit: 5000`
   - `X-Ratelimit-Remaining: 4983`
   - `X-Ratelimit-Reset: 1746963906`
   - `X-Ratelimit-Used: 17`
   - `X-Ratelimit-Resource: core`

3. **安全头部**：
   - `Access-Control-Expose-Headers`: 控制浏览器可访问的响应头

### 自定义配置
如需修改默认配置，可以直接编辑 `resp.go` 文件中的 `SetHeader` 函数。

## 🧪 测试建议

### 单元测试要点
1. 测试 `HandleSuccess` 返回正确的 HTTP 状态码和格式
2. 测试 `HandleError` 在不同状态码下的行为
3. 测试 `GeneratePaginationLinks` 生成正确的分页链接
4. 验证 HTTP 头部设置是否正确

### 集成测试要点
1. 验证完整的 API 响应流程
2. 测试 CORS 头部在实际请求中的效果
3. 验证分页功能与前端客户端的兼容性

## 📚 相关文档

- [PieMDM 目录结构分析](../../../../docs/directory-structure-analysis.md)
- [PieMDM 项目结构文档](../../../../docs/project-structure.md)
- [Go Gin 框架文档](https://gin-gonic.com/docs/)
- [GitHub API 速率限制文档](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)

## 🏷️ 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.0 | 2024-01 | 初始版本，包含基本响应处理功能 |
| 1.1 | 2024-01 | 添加分页链接生成功能 |

---

**维护团队**: 技术架构组
**最后更新**: 2024年1月
**状态**: 活跃维护中 ✅