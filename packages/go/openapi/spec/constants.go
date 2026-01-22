package spec

import "time"

// Header 常量定义
const (
	HeaderAppID     = "X-App-Id"
	HeaderTimestamp = "X-Timestamp"
	HeaderNonce     = "X-Nonce"
	HeaderSignature = "X-Sign"
	HeaderschemaVer = "X-Schema-Version"
)

// Default Config
const (
	DefaultTimestampWindow = 5 * time.Minute
	DefaultNonceTTL        = 10 * time.Minute
	EmptyPayloadHash       = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // SHA256("")
)

// SignOptions 签名配置选项
type SignOptions struct {
	// 允许自定义 Header 名称，留空则使用默认值
	AppIDHeader     string
	TimestampHeader string
	NonceHeader     string
	SignatureHeader string
}

func (o SignOptions) GetAppIDHeader() string {
	if o.AppIDHeader != "" {
		return o.AppIDHeader
	}
	return HeaderAppID
}

func (o SignOptions) GetTimestampHeader() string {
	if o.TimestampHeader != "" {
		return o.TimestampHeader
	}
	return HeaderTimestamp
}

func (o SignOptions) GetNonceHeader() string {
	if o.NonceHeader != "" {
		return o.NonceHeader
	}
	return HeaderNonce
}

func (o SignOptions) GetSignatureHeader() string {
	if o.SignatureHeader != "" {
		return o.SignatureHeader
	}
	return HeaderSignature
}
