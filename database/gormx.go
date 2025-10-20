package database

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqlDatabase struct {
	db  *gorm.DB
	cfg SqlDbConfig
}

var (
	databaseInstace *SqlDatabase
	once            sync.Once
)

func NewDatabase(conf SqlDbConfig, engine string) IDatabase {
	once.Do(func() {
		var dsn string
		var err error
		var conn *gorm.DB

		// Switch on the engine to determine which database to connect to
		switch engine {
		case "postgres":
			dsn = fmt.Sprintf(
				"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s search_path=%s",
				conf.Host,
				conf.User,
				conf.Password,
				conf.DBName,
				conf.Port,
				conf.SSLMode,
				conf.Schema,
			)
			conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		case "mysql":
			dsn = fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
				conf.User,
				conf.Password,
				conf.Host,
				conf.Port,
				conf.DBName,
			)
			conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		case "sqlite":
			dsn = conf.DBName // SQLite uses the DB file path as the dsn
			conn, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		default:
			log.Fatalf("Unsupported database engine: %s", engine)
		}

		if err != nil {
			panic(err)
		}

		log.Printf("Connected to database %s", conf.DBName)

		databaseInstace = &SqlDatabase{
			db: conn,
		}
	})

	return databaseInstace
}

func (db *SqlDatabase) Connect() *gorm.DB {
	return databaseInstace.db
}
