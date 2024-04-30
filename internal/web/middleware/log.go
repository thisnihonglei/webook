package middleware

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type LogMiddlewareBuilder struct {
	logFunc       func(ctx context.Context, al AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddlewareBuilder(logFunc func(ctx context.Context, al AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFunc: logFunc,
	}
}

func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

type AccessLog struct {
	Method   string        `json:"method"`
	Path     string        `json:"path"`
	ReqBody  string        `json:"req_body"`
	RespBody string        `json:"resp_body"`
	Duration time.Duration `json:"duration"`
	Status   int           `json:"status"`
}

func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) > 1024 {
			path = path[:1024]
		}
		method := c.Request.Method

		al := AccessLog{
			Method: method,
			Path:   path,
		}
		if l.allowReqBody {
			// Request.Body 是stream对象，只能读一次
			body, _ := c.GetRawData()
			if len(body) > 2048 {
				al.ReqBody = string(body[:2048])
			} else {
				al.ReqBody = string(body)
			}
			// 放回Request.Body
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		startTime := time.Now()
		if l.allowRespBody {
			c.Writer = &responseWriter{
				ResponseWriter: c.Writer,
				al:             &al,
			}
		}

		defer func() {
			duration := time.Since(startTime)
			al.Duration = duration
			l.logFunc(c, al)
		}()

		c.Next()
	}
}

type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	w.al.RespBody = s
	return w.ResponseWriter.WriteString(s)
}

func (w *responseWriter) WriteHeader(code int) {
	w.al.Status = code
	w.ResponseWriter.WriteHeader(code)
}
