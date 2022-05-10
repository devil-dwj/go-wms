package log

import (
	"time"

	"go.uber.org/zap"
	"moul.io/zapgorm2"
)

type GormLog struct {
	zapgorm2.Logger
}

func NewGormLog(l *zap.Logger) *GormLog {
	return &GormLog{
		Logger: zapgorm2.New(l),
	}
}

func (l *GormLog) SlowHold(t time.Duration) {
	l.Logger.SlowThreshold = t
}
