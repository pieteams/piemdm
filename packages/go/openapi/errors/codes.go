package errors

import "net/http"

// 定义标准错误码
// 使用 var 而不是 const，允许外部库扩展或覆盖（虽然不建议覆盖标准码）
var (
	// Auth Errors
	ErrAuthFailed       = &openApiError{"AUTH_FAILED", http.StatusUnauthorized, "Authentication failed"}
	ErrSignatureInvalid = &openApiError{"AUTH_SIGNATURE_INVALID", http.StatusUnauthorized, "Signature mismatch"}
	ErrTokenExpired     = &openApiError{"AUTH_TOKEN_EXPIRED", http.StatusUnauthorized, "Timestamp expired"}
	ErrNonceUsed        = &openApiError{"AUTH_NONCE_USED", http.StatusUnauthorized, "Nonce already used"}
	ErrIpNotAllowed     = &openApiError{"AUTH_IP_NOT_ALLOWED", http.StatusForbidden, "IP not in whitelist"}

	// Permission Errors
	ErrPermissionDenied = &openApiError{"PERMISSION_DENIED", http.StatusForbidden, "Permission denied"}

	// Rate Limit Errors
	ErrRateLimitExceeded = &openApiError{"RATE_LIMIT_EXCEEDED", http.StatusTooManyRequests, "Rate limit exceeded"}

	// Param Errors
	ErrParamMissing = &openApiError{"PARAM_REQUIRED_MISSING", http.StatusBadRequest, "Missing required parameter"}
	ErrParamInvalid = &openApiError{"PARAM_VALUE_INVALID", http.StatusBadRequest, "Invalid parameter value"}

	// System Errors
	ErrSystemError = &openApiError{"SYSTEM_INTERNAL_ERROR", http.StatusInternalServerError, "Internal system error"}
)
