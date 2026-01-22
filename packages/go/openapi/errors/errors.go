package errors

import "fmt"

// ErrorCode 错误码接口
type ErrorCode interface {
	Code() string
	HTTPStatus() int
	Message() string
}

// openApiError 实现 ErrorCode 接口
type openApiError struct {
	code       string
	httpStatus int
	message    string
}

func (e *openApiError) Code() string {
	return e.code
}

func (e *openApiError) HTTPStatus() int {
	return e.httpStatus
}

func (e *openApiError) Message() string {
	return e.message
}

func (e *openApiError) Error() string {
	return fmt.Sprintf("[%s] %s", e.code, e.message)
}

// New 创建一个新的错误
func New(code string, httpStatus int, message string) error {
	return &openApiError{
		code:       code,
		httpStatus: httpStatus,
		message:    message,
	}
}

// Is 判断是否为指定错误码 (仅匹配Code)
func Is(err error, target ErrorCode) bool {
	if e, ok := err.(ErrorCode); ok {
		return e.Code() == target.Code()
	}
	return false
}
