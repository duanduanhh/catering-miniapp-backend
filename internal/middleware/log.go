package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

func RequestLogMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// The configuration is initialized once per request
		uuid, err := random.UUIdV4()
		if err != nil {
			return
		}
		trace := cryptor.Md5String(uuid)
		ctx.Set("trace", trace) // 设置到gin.Context
		logger.WithValue(ctx, zap.String("trace", trace))
		logger.WithValue(ctx, zap.String("request_method", ctx.Request.Method))
		logger.WithValue(ctx, zap.Any("request_headers", ctx.Request.Header))
		logger.WithValue(ctx, zap.String("request_url", ctx.Request.URL.String()))
		if ctx.Request.Body != nil {
			contentType := ctx.GetHeader("Content-Type")
			if !isTextRequestBody(contentType) {
				logger.WithValue(ctx, zap.String("request_params", "<body omitted>"))
			} else {
				bodyBytes, _ := ctx.GetRawData()
				ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 关键点

				// Format JSON request body for better readability
				var formattedBody any
				if len(bodyBytes) > 0 {
					if err := json.Unmarshal(bodyBytes, &formattedBody); err == nil {
						logger.WithValue(ctx, zap.Any("request_params", formattedBody))
					} else {
						logger.WithValue(ctx, zap.String("request_params", string(bodyBytes)))
					}
				}
			}
		}
		logger.WithContext(ctx).Info("Request")
		ctx.Next()
	}
}

func ResponseLogMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		startTime := time.Now()
		ctx.Next()
		duration := time.Since(startTime).String()

		// Format JSON response body for better readability
		responseBody := blw.body.String()
		var formattedBody any
		if err := json.Unmarshal([]byte(responseBody), &formattedBody); err == nil {
			logger.WithContext(ctx).Info("Response", zap.Any("response_body", formattedBody), zap.Any("time", duration))
		} else {
			// Fallback to string if JSON parsing fails
			logger.WithContext(ctx).Info("Response", zap.String("response_body", responseBody), zap.Any("time", duration))
		}
	}
}

func isTextRequestBody(contentType string) bool {
	if contentType == "" {
		return true
	}
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/") ||
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "+json")
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
