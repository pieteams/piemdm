package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pieworks/piemdm/packages/go/openapi/auth"
	"github.com/pieworks/piemdm/packages/go/openapi/spec"
)

// Client OpenAPI 客户端
type Client struct {
	baseURL     string
	appID       string
	appSecret   string
	httpClient  *http.Client
	signOptions spec.SignOptions
}

// Config 客户端配置
type Config struct {
	BaseURL     string
	AppID       string
	AppSecret   string
	Timeout     time.Duration
	SignOptions spec.SignOptions
}

// NewClient 创建新的客户端
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Client{
		baseURL:   cfg.BaseURL,
		appID:     cfg.AppID,
		appSecret: cfg.AppSecret,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		signOptions: cfg.SignOptions,
	}
}

// Do 发送请求
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// 1. 准备签名参数
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := uuid.New().String()

	// 2. 读取 Body 用于计算哈希 (如果 Body 不为空)
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 恢复 Body
	}

	// 3. 构建 Canonical Request
	canonicalRequest := auth.BuildCanonicalRequest(
		req.Method,
		req.URL.Path,
		req.URL.Query(),
		bodyBytes,
		timestamp,
		nonce,
	)

	// 4. 计算签名
	signature := auth.ComputeSignature(canonicalRequest, c.appSecret)

	// 5. 设置 Headers
	req.Header.Set(c.signOptions.GetAppIDHeader(), c.appID)
	req.Header.Set(c.signOptions.GetTimestampHeader(), timestamp)
	req.Header.Set(c.signOptions.GetNonceHeader(), nonce)
	req.Header.Set(c.signOptions.GetSignatureHeader(), signature)
	req.Header.Set("Content-Type", "application/json") // 默认 JSON

	// 6. 发送请求
	return c.httpClient.Do(req)
}

// Get 发送 GET 请求辅助方法
func (c *Client) Get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post 发送 POST 请求辅助方法
func (c *Client) Post(path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
