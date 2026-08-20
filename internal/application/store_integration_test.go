package application

import (
	"context"
	"os"
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

	// 4. Create a Store on the deployed database
	store, err := NewStore(dsn)
	if err != nil {
		t.Fatalf("couldn't create store: %v", err)
	}

	// 5. Apply the schema — create the `applications` table
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatalf("couldn't read schema: %v", err)
	}

	if _, err := store.db.ExecContext(ctx, string(schema)); err != nil {
		t.Fatalf("couldn't apply schema: %v", err)
	}

	// 6. Save the request via Add
	saved, err := store.Add(Application{Client: "Alina", Amount: 1000, Term: 12})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	// 7. Check what was returned
	if saved.ID == 0 { // WHY: Checking the property ("definition ID, it isn't empty")
		t.Errorf("expected a non-zero id from the database, got %d", saved.ID)
	}

	if saved.Status != "new" {
		t.Errorf("expected status 'new', got %q", saved.Status)
	}

	// 8.  Verify that the record actually exists in the database
	var (
		client string
		amount int
		status string
	)
	
	err = store.db.QueryRowContext(ctx,
		"SELECT client, amount, status FROM applications WHERE id = $1",
		saved.ID,
	).Scan(&client, &amount, &status)

	// WHY: This proves that the data was actually written, rather than just returned from the function
	if err != nil {
		t.Fatalf("couldn't read application back: %v", err)
	}

	if client != "Alina" || amount != 1000 || status != "new" {
		t.Errorf("data in db mismatch: got client=%q amount=%d status=%q", client, amount, status)
	}
}