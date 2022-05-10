package mysql

import (
	"time"

	"github.com/devil-dwj/go-wms/base/log"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func GetDB(dsn string, l *zap.Logger) *gorm.DB {
	zl := gormLog(l)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: zl,
	})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxIdleConns(8)                   // 数据库连接数
	sqlDB.SetMaxOpenConns(4)                   // 连接池最大空闲连接数
	sqlDB.SetConnMaxIdleTime(time.Second * 30) // 连接池里的连接最大空闲时长，超时会被清理
	sqlDB.SetConnMaxLifetime(time.Minute * 10) // 连接的最大时长

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		panic(err)
	}
	return db
}

func gormLog(l *zap.Logger) gormlogger.Interface {
	zl := log.NewGormLog(l)
	zl.LogMode(gormlogger.Silent)
	zl.SlowHold(time.Millisecond * 100) // 数据库反应时间, 超过打印日志

	return zl
}
