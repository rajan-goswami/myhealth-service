// Package postgres implements postgres connection.
package postgres

import (
	"fmt"
	"log"
	"myhealth-service/internal/entity"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	_defaultMaxPoolSize     = 1
	_defaultConnAttempts    = 10
	_defaultConnTimeout     = time.Second
	_defaultConnMaxLifetime = time.Hour
	_defaultDisableLogger   = false
)

// Postgres -.
type Postgres struct {
	maxPoolSize     int
	connAttempts    int
	connTimeout     time.Duration
	connMaxLifetime time.Duration
	disableLogger   bool

	db *gorm.DB
}

// New -.
func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:     _defaultMaxPoolSize,
		connAttempts:    _defaultConnAttempts,
		connTimeout:     _defaultConnTimeout,
		connMaxLifetime: _defaultConnMaxLifetime,
		disableLogger:   _defaultDisableLogger,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	var err error
	for pg.connAttempts > 0 {
		gormConfig := &gorm.Config{}
		if pg.disableLogger {
			gormConfig = &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			}
		}
		pg.db, err = gorm.Open(postgres.Open(url), gormConfig)
		if err != nil {
			log.Printf("unable to open database connection: %v\n", err)
		}
		// ping database
		sqlDB, err := pg.db.DB()
		if err != nil {
			log.Printf("unable to access database configurations: %v\n", err)
		}
		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		sqlDB.SetMaxIdleConns(pg.maxPoolSize)
		// SetMaxOpenConns sets the maximum number of open connections to the database.
		sqlDB.SetMaxOpenConns(pg.maxPoolSize)
		// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		sqlDB.SetConnMaxLifetime(pg.connMaxLifetime)
		// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
		sqlDB.SetConnMaxIdleTime(pg.connMaxLifetime * 5)
		err = sqlDB.Ping()
		if err != nil {
			log.Printf("unable to open database connection: %v\n", err)
		}
		// no error generated, connection is successful
		if err == nil {
			log.Println("connected successfully to database")
			pg.db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
			break
		}
		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)
		time.Sleep(pg.connTimeout)
		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - connAttempts == 0: %w", err)
	}

	return pg, nil
}

// This migrate all tables
func AutoMigrate(pg *gorm.DB) error {
	return pg.AutoMigrate(
		&entity.GlucoseRecord{},
		&entity.GlucoseRecordSyncMeta{},
	)
}

func (pg *Postgres) Database() *gorm.DB {
	return pg.db
}

// Ping checks if postgres connection is alive
func (pg *Postgres) Ping() error {
	databaseConn, err := pg.Database().DB()
	if err != nil {
		return err
	}
	return databaseConn.Ping()
}

// Close -.
func (pg *Postgres) Close() error {
	databaseConn, err := pg.db.DB()
	if err != nil {
		return err
	}
	return databaseConn.Close()
}
