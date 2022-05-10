package log

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func MustLog(logname string) *zap.Logger {
	w := getWriter(logname)
	muw := zapcore.NewMultiWriteSyncer(w, os.Stdout)
	e := getEncoder()

	c := zapcore.NewCore(e, muw, zapcore.DebugLevel)
	l := zap.New(c, zap.AddCaller())

	zap.ReplaceGlobals(l)

	return l
}

func getWriter(name string) zapcore.WriteSyncer {
	l := &lumberjack.Logger{
		Filename:   name,
		MaxSize:    500,
		MaxBackups: 500,
		MaxAge:     30,
		LocalTime:  true,
		Compress:   false,
	}
	return zapcore.AddSync(l)
}

func getEncoder() zapcore.Encoder {
	e := zap.NewProductionEncoderConfig()
	e.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	e.TimeKey = "time"
	e.EncodeLevel = zapcore.CapitalLevelEncoder
	e.EncodeDuration = zapcore.MillisDurationEncoder
	e.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewConsoleEncoder(e)
}
