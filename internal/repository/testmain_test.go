package repository_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
)

var testPG *pgxdriver.Postgres

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		panic("start postgres container: " + err.Error())
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("get connection string: " + err.Error())
	}

	testPG, err = pgxdriver.New(dsn, noopLogger{})
	if err != nil {
		panic("connect to postgres: " + err.Error())
	}

	mig, err := migrate.New("file://../../migrations", dsn)
	if err != nil {
		panic("create migrator: " + err.Error())
	}
	if err = mig.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic("run migrations: " + err.Error())
	}

	code := m.Run()

	testPG.Close()
	_ = pgContainer.Terminate(ctx)

	os.Exit(code)
}
