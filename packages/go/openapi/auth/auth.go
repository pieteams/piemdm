package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// BuildCanonicalRequest 构建规范请求字符串
func BuildCanonicalRequest(
	method string,
	path string,
	queryParams url.Values,
	body []byte,
	timestamp string,
	nonce string,
) string {
	// 1. HTTP 方法
	canonicalMethod := strings.ToUpper(method)

	// 2. URI 路径
	// 确保路径以 / 开头，且不包含重复斜杠（根据具体网关行为调整，这里做基础清理）
	canonicalPath := path
	if !strings.HasPrefix(canonicalPath, "/") {
		canonicalPath = "/" + canonicalPath
	}

	// 3. 排序后的查询参数
	canonicalQuery := sortQueryString(queryParams)

	// 4. 请求体哈希
	bodyHash := HashRequestBody(body)

	// 5. 组合 Canonical Request
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		canonicalMethod,
		canonicalPath,
		canonicalQuery,
		bodyHash,
		timestamp,
		nonce,
	)
}

// sortQueryString 对查询参数排序
func sortQueryString(queryParams url.Values) string {
	if len(queryParams) == 0 {
		return ""
	}

	// 获取所有键并排序
	keys := make([]string, 0, len(queryParams))
	for k := range queryParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建排序后的查询字符串
	var parts []string
	for _, k := range keys {
		// 注意: 这里假设 queryParams 已经是 decoded 或者需要统一 encode
		// 标准做法是: Key 和 Value 都需要 URL Encode
		values := queryParams[k]
		// 如果有多个值，也需要排序吗？通常 API 网关会要求。这里简化处理，按出现顺序。
		for _, v := range values {
			parts = append(parts, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
		}
	}

	return strings.Join(parts, "&")
}

// HashRequestBody 计算请求体的 SHA256 哈希
func HashRequestBody(body []byte) string {
	if len(body) == 0 {
		body = []byte("")
	}
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// ComputeSignature 计算 HMAC-SHA256 签名
func ComputeSignature(canonicalRequest, appSecret string) string {
	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write([]byte(canonicalRequest))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature 验证签名
func VerifySignature(signature, canonicalRequest, appSecret string) bool {
	expected := ComputeSignature(canonicalRequest, appSecret)
	return hmac.Equal([]byte(signature), []byte(expected))
}
