package postgres

import (
	"database/sql"
	"log"
	"os"

	"github.com/catalystsquad/app-utils-go/env"
	"github.com/catalystsquad/app-utils-go/logging"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresStorage struct{}

var uri = env.GetEnvOrDefault("POSTGRES_URI", "postgresql://postgres:postgres@localhost:5432?sslmode=disable")
var gormLogLevel = env.GetEnvOrDefault("POSTGRES_LOG_LEVEL", "error")
var ignoreRecordNotFoundError = env.GetEnvAsBoolOrDefault("POSTGRES_LOG_IGNORE_NOT_FOUND", "true")
var parameterizedQueries = env.GetEnvAsBoolOrDefault("POSTGRES_LOG_PARAMETERIZED_QUERIES", "true")
var slowQueryThreshold = env.GetEnvAsDurationOrDefault("POSTGRES_LOG_SLOW_QUERY_THRESHOLD", "1s")
var maxIdleConns = env.GetEnvAsIntOrDefault("POSTGRES_MAX_IDLE_CONNS", "10")
var maxOpenConns = env.GetEnvAsIntOrDefault("POSTGRES_MAX_OPEN_CONNS", "100")
var connMaxLifetime = env.GetEnvAsDurationOrDefault("POSTGRES_CONN_MAX_LIFETIME", "1h")
var connMaxIdleTime = env.GetEnvAsDurationOrDefault("POSTGRES_CONN_MAX_IDLE_TIME", "10m")
var db *gorm.DB

func (p PostgresStorage) Initialize() (shutdown func(), err error) {
	config := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             slowQueryThreshold,
				ParameterizedQueries:      parameterizedQueries,
				LogLevel:                  getGormLogLevel(),         // Log level
				IgnoreRecordNotFoundError: ignoreRecordNotFoundError, // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,                      // Disable color
			},
		),
	}
	if db, err = gorm.Open(postgres.Open(uri), config); err != nil {
		return nil, err
	}
	var sqlDb *sql.DB
	if sqlDb, err = db.DB(); err != nil {
		return nil, err
	}
	sqlDb.SetMaxIdleConns(maxIdleConns)
	sqlDb.SetMaxOpenConns(maxOpenConns)
	sqlDb.SetConnMaxLifetime(connMaxLifetime)
	sqlDb.SetConnMaxIdleTime(connMaxIdleTime)
	logging.Log.Info("connected to postgres")
	//err = db.AutoMigrate(&katanov1.AgentGormModel{})
	//if err != nil {
	//	panic(err)
	//}
	return nil, nil
}

func getGormLogLevel() logger.LogLevel {
	switch gormLogLevel {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Error
	}
}

func (p PostgresStorage) Ready() bool {
	sqlDb, err := db.DB()
	if err != nil {
		return false
	}
	return sqlDb.Ping() == nil
}
