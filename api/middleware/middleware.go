package middleware

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterMiddleWares(r *gin.Engine, middleware ...gin.HandlerFunc) {
	r.Use(middleware...)
}

func WithLogger(l *zap.Logger) gin.HandlerFunc {
	return logger(l)
}

func WithRecovery(stack bool) gin.HandlerFunc {
	return recovery(stack)
}

func WithCors() gin.HandlerFunc {
	return mcors()
}

func mcors() gin.HandlerFunc {
	return cors.New(
		cors.Config{
			AllowAllOrigins: true,
			//AllowOrigins:    nil,
			AllowMethods: []string{
				"OPTIONS",
				"GET",
				"POST",
				"PUT",
				"PATCH",
				"DELETE",
				"FETCH",
			},
			AllowHeaders:           []string{"Authorization, Content-Length, X-CSRF-Token, Token,session", "Content-Type", "x-requested-with"},
			AllowCredentials:       true,
			ExposeHeaders:          []string{"Content-Length", "Content-Type"},
			MaxAge:                 86400,
			AllowWildcard:          true,
			AllowBrowserExtensions: true,
			AllowWebSockets:        true,
			AllowFiles:             true,
		},
	)
}

func logger(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		l.Info(
			path,
			zap.Int("status", c.Writer.Status()),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			// zap.String("user-agent", c.Request.UserAgent()),
			zap.String("error", c.Errors.String()),
			zap.Duration("cost", cost),
		)
	}
}

func recovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zap.L().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

type MiddleWareRecord struct {
	Logger  *zap.Logger
	Request *http.Request
	Start   time.Time
	Status  int
	Err     string
	Cost    time.Duration
}

func (r *MiddleWareRecord) Log() {
	r.Logger.Info(
		r.Request.URL.Path,
		zap.Int("status", r.Status),
		zap.String("path", r.Request.URL.Path),
		zap.String("query", r.Request.URL.RawQuery),
		zap.String("host", r.Request.Host),
		zap.String("error", r.Err),
		zap.Duration("cost", r.Cost),
	)
}

func (r *MiddleWareRecord) LogRecovery(err interface{}) {
	r.Logger.Error(
		"[Recovery from panic]"+r.Request.URL.Path,
		zap.Any("error", err),
	)
}

func Logger(record *MiddleWareRecord) {
	record.Log()
}

func Recovery(record *MiddleWareRecord) {
	defer func() {
		if err := recover(); err != nil {
			record.LogRecovery(err)
		}
	}()
}
