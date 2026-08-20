package application

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestStoreAddIntegration(t *testing.T) {
	ctx := context.Background()

	// 1. Spin up Postgres in a container
	pgContainer, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("credit"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("secret"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(30*time.Second), // WHY: Wait until port 5432 actually starts listening
		),
	)

	if err != nil {
		t.Fatalf("couldn't start postgres container: %v", err)
	}

	// 2. Ensure the container is shut down after the test
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("couldn't terminate container: %v", err)
		}
	})

	// 3. Get the connection string for the running container
	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable") // WHY: Testcontainers automatically detects the port on which the container has started
	if err != nil {
		t.Fatalf("couldn't get connection string: %v", err)
	}

	t.Logf("postgres is up at: %s", dsn)
}