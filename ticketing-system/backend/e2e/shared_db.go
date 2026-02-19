//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type sharedDBState struct {
	container *postgres.PostgresContainer
	connStr   string
	adminPool *pgxpool.Pool
}

var sharedDB *sharedDBState

func initSharedDB(ctx context.Context) error {
	container, err := runPostgresContainer(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ticketing"),
		postgres.WithUsername("ticketing"),
		postgres.WithPassword("ticketing"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(90*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("start shared postgres container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx) //nolint:errcheck
		return fmt.Errorf("shared postgres connection string: %w", err)
	}

	adminPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		container.Terminate(ctx) //nolint:errcheck
		return fmt.Errorf("shared postgres admin pool: %w", err)
	}

	// Pre-install extensions to avoid concurrent migration races.
	if _, err := adminPool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "pgcrypto"`); err != nil {
		adminPool.Close()
		container.Terminate(ctx) //nolint:errcheck
		return fmt.Errorf("install pgcrypto: %w", err)
	}

	sharedDB = &sharedDBState{
		container: container,
		connStr:   connStr,
		adminPool: adminPool,
	}
	return nil
}

func cleanupSharedDB() {
	if sharedDB == nil {
		return
	}
	if sharedDB.adminPool != nil {
		sharedDB.adminPool.Close()
	}
	if sharedDB.container != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		sharedDB.container.Terminate(ctx) //nolint:errcheck
	}
	sharedDB = nil
}

func createTestSchema(ctx context.Context, name string) error {
	_, err := sharedDB.adminPool.Exec(ctx, "CREATE SCHEMA "+name)
	return err
}

func dropTestSchema(ctx context.Context, name string) error {
	_, err := sharedDB.adminPool.Exec(ctx, "DROP SCHEMA IF EXISTS "+name+" CASCADE")
	return err
}

func newTestStore(ctx context.Context, schemaName string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(sharedDB.connStr)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schemaName + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool for schema %s: %w", schemaName, err)
	}
	return pool, nil
}

func runPostgresContainer(ctx context.Context, image string, opts ...testcontainers.ContainerCustomizer) (container *postgres.PostgresContainer, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("testcontainers panic: %v", r)
		}
	}()
	return postgres.Run(ctx, image, opts...)
}
