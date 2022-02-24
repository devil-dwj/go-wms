package mysql

import (
	"time"

	"github.com/devil-dwj/go-wms/log"
	"go.uber.org/zap"
	mysqlraw "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var WmsDB *gorm.DB

func WmsMysql(dsn string, l *zap.Logger) error {

	// zl := zapgorm2.New(l)
	zl := log.NewGormLog(l)
	zl.LogMode(gormlogger.Silent)
	zl.SlowHold(time.Second)

	db, err := gorm.Open(mysqlraw.Open(dsn), &gorm.Config{
		Logger: zl,
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(8)                   // 数据库连接数
	sqlDB.SetMaxOpenConns(4)                   // 连接池最大空闲连接数
	sqlDB.SetConnMaxIdleTime(time.Second * 30) // 连接池里的连接最大空闲时长，超时会被清理
	sqlDB.SetConnMaxLifetime(time.Minute * 10) // 连接的最大时长

	WmsDB = db

	return nil
}
