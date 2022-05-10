package middleware

import (
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type MiddleWareRecord struct {
	Logger  *zap.Logger
	Tracer  opentracing.Tracer
	Request *http.Request
	Start   time.Time
	Status  int
	Err     interface{}
	Cost    time.Duration
}

func (r *MiddleWareRecord) Log() {
	r.Logger.Info(
		r.Request.URL.Path,
		zap.Int("status", r.Status),
		zap.String("path", r.Request.URL.Path),
		zap.String("query", r.Request.URL.RawQuery),
		zap.String("host", r.Request.Host),
		zap.Any("error", r.Err),
		zap.Duration("cost", r.Cost),
	)
}

func (r *MiddleWareRecord) LogRecovery() {
	r.Logger.Error(
		"[Recovery from panic]"+r.Request.URL.Path,
		zap.Any("error", r.Err),
	)
}

func Logger(record *MiddleWareRecord) error {
	record.Log()
	return nil
}

func Recovery(record *MiddleWareRecord) error {
	record.LogRecovery()
	return nil
}

func Tracing(record *MiddleWareRecord) error {

	return nil
}
