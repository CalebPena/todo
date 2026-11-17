package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"todo/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

var dbPort string = "5432"
var dbHost string = "127.0.0.1"
var dbDriverName = "postgres"

const dbKey string = "database"

//go:embed migrations/*.sql
var migrationFiles embed.FS

func InitDb() (*sql.DB, error) {
	host := os.Getenv("POSTGRES_HOST")
	if host != "" {
		dbHost = host
	}
	port := os.Getenv("POSTGRES_PORT")
	if port != "" {
		dbPort = port
	}

	sslMode := "require"
	if util.IsDebug {
		sslMode = "disable"
	}
	if configured := os.Getenv("POSTGRES_SSLMODE"); configured != "" {
		sslMode = configured
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPort,
			util.GetEnvironmentVariable("POSTGRES_USER"), util.GetEnvironmentVariable("POSTGRES_PASSWORD"), util.GetEnvironmentVariable("POSTGRES_DB"), sslMode)
	}

	db, err := sql.Open(dbDriverName, connStr)

	if err != nil {
		return nil, err
	}

	instance, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	source, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return nil, err
	}
	migrations, err := migrate.NewWithInstance("iofs", source, dbDriverName, instance)
	if err != nil {
		return nil, err
	}

	err = migrations.Up()

	if err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return db, nil
}

func DbMiddleware(db *sql.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		c.Set(dbKey, db)
	}
}

func GetDb(c *gin.Context) *sql.DB {
	return c.MustGet(dbKey).(*sql.DB)
}
