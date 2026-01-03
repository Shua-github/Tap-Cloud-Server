package types

import "net/http"

type Code int

// 自定义
const (
	// 数据库
	DbError Code = iota
	DbNotFound
	// 签名
	BadSign
)

type TCSError struct {
	HTTPCode int    `json:"code"`
	TCSCode  Code   `json:"_code,omitempty"`
	Message  string `json:"error,omitempty"`
}

// 默认 HTTP 错误
var BadRequestError = TCSError{
	HTTPCode: http.StatusBadRequest,
	Message:  http.StatusText(http.StatusBadRequest),
}

var MethodNotAllowedError = TCSError{
	HTTPCode: http.StatusMethodNotAllowed,
	Message:  http.StatusText(http.StatusMethodNotAllowed),
}

var NotFoundError = TCSError{
	HTTPCode: http.StatusNotFound,
	Message:  http.StatusText(http.StatusNotFound),
}

var UnauthorizedError = TCSError{
	HTTPCode: http.StatusUnauthorized,
	Message:  http.StatusText(http.StatusUnauthorized),
}

var ForbiddenError = TCSError{
	HTTPCode: http.StatusForbidden,
	Message:  http.StatusText(http.StatusForbidden),
}

func NewUnknownError(msg string) TCSError {
	return TCSError{
		HTTPCode: 500,
		Message:  msg,
	}
}
