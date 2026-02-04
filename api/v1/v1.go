package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
	Data    any    `json:"data"`
}

func getTraceID(ctx *gin.Context) string {
	if trace, exists := ctx.Get("trace"); exists {
		if traceID, ok := trace.(string); ok {
			return traceID
		}
	}
	return ""
}

func HandleSuccess(ctx *gin.Context, data any) {
	if data == nil {
		data = map[string]any{}
	}
	resp := Response{
		Code:    errorCodeMap[ErrSuccess],
		Message: ErrSuccess.Error(),
		Data:    data,
		TraceID: getTraceID(ctx),
	}
	if _, ok := errorCodeMap[ErrSuccess]; !ok {
		resp = Response{Code: 0, Message: "", Data: data, TraceID: getTraceID(ctx)}
	}
	ctx.JSON(http.StatusOK, resp)
}

func HandleError(ctx *gin.Context, httpCode int, err error, data any) {
	if data == nil {
		data = map[string]string{}
	}
	resp := Response{
		Code:    errorCodeMap[err],
		Message: err.Error(),
		Data:    data,
		TraceID: getTraceID(ctx),
	}
	if _, ok := errorCodeMap[err]; !ok {
		resp = Response{Code: 500, Message: "unknown error", Data: data, TraceID: getTraceID(ctx)}
	}
	ctx.JSON(httpCode, resp)
}

type Error struct {
	Code    int
	Message string
}

var errorCodeMap = map[error]int{}

func newError(code int, msg string) error {
	err := errors.New(msg)
	errorCodeMap[err] = code
	return err
}

func (e Error) Error() string {
	return e.Message
}
