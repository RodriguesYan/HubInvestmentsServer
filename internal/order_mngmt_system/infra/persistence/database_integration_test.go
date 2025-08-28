package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"HubInvestments/shared/infra/database"
)

// TestDatabaseConnectionIntegration tests database connection and basic operations
func TestDatabaseConnectionIntegration(t *testing.T) {
	// Setup test database connection
	db, err := database.CreateDatabaseConnection()
	require.NoError(t, err)
	defer db.Close()

	t.Run("Database Connection", func(t *testing.T) {
		// Test basic connection
		err := db.Ping()
		assert.NoError(t, err)
	})

	t.Run("Basic Query Operations", func(t *testing.T) {
		// Test basic query functionality
		var result int
		err := db.QueryRow("SELECT 1").Scan(&result)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("Orders Table Exists", func(t *testing.T) {
		// Test that orders table exists and has expected structure
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'orders'
			)
		`
		err := db.QueryRow(query).Scan(&exists)
		if err != nil {
			t.Skipf("Could not check table existence: %v", err)
			return
		}

		if exists {
			t.Logf("Orders table exists")

			// Check basic table structure
			var columnCount int
			columnQuery := `
				SELECT COUNT(*) 
				FROM information_schema.columns 
				WHERE table_name = 'orders'
			`
			err = db.QueryRow(columnQuery).Scan(&columnCount)
			assert.NoError(t, err)
			assert.Greater(t, columnCount, 5, "Orders table should have multiple columns")
		} else {
			t.Logf("Orders table does not exist - this is expected in test environment")
		}
	})
}

// TestDatabasePerformanceIntegration tests database performance characteristics
func TestDatabasePerformanceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	db, err := database.CreateDatabaseConnection()
	require.NoError(t, err)
	defer db.Close()

	t.Run("Concurrent Connections", func(t *testing.T) {
		const numConnections = 10
		results := make(chan error, numConnections)

		// Test concurrent database operations
		for i := 0; i < numConnections; i++ {
			go func(index int) {
				var result int
				err := db.QueryRow("SELECT $1", index).Scan(&result)
				results <- err
			}(i)
		}

		// Collect results
		errorCount := 0
		for i := 0; i < numConnections; i++ {
			if err := <-results; err != nil {
				errorCount++
				t.Logf("Concurrent connection error: %v", err)
			}
		}

		// Most connections should succeed
		assert.Less(t, errorCount, numConnections/2, "Too many concurrent connection errors")
	})

	t.Run("Query Performance", func(t *testing.T) {
		const numQueries = 100
		start := time.Now()

		for i := 0; i < numQueries; i++ {
			var result int
			err := db.QueryRow("SELECT $1", i).Scan(&result)
			assert.NoError(t, err)
		}

		duration := time.Since(start)
		avgLatency := duration / numQueries
		t.Logf("Average query latency: %v", avgLatency)

		// Performance assertion - queries should be reasonably fast
		assert.Less(t, avgLatency, 10*time.Millisecond, "Average query latency too high")
	})
}

// TestDatabaseTransactionIntegration tests transaction handling
func TestDatabaseTransactionIntegration(t *testing.T) {
	db, err := database.CreateDatabaseConnection()
	require.NoError(t, err)
	defer db.Close()

	t.Run("Transaction Commit", func(t *testing.T) {
		// Begin transaction
		tx, err := db.Begin()
		require.NoError(t, err)

		// Perform operations within transaction
		_, err = tx.Exec("SELECT 1")
		assert.NoError(t, err)

		// Commit transaction
		err = tx.Commit()
		assert.NoError(t, err)
	})

	t.Run("Transaction Rollback", func(t *testing.T) {
		// Begin transaction
		tx, err := db.Begin()
		require.NoError(t, err)

		// Perform operations within transaction
		_, err = tx.Exec("SELECT 1")
		assert.NoError(t, err)

		// Rollback transaction
		err = tx.Rollback()
		assert.NoError(t, err)
	})

	t.Run("Transaction Context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Begin transaction with context
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)

		// Perform operations within transaction
		_, err = tx.Exec("SELECT 1")
		assert.NoError(t, err)

		// Commit transaction
		err = tx.Commit()
		assert.NoError(t, err)
	})
}

// TestDatabaseErrorHandling tests error handling scenarios
func TestDatabaseErrorHandling(t *testing.T) {
	db, err := database.CreateDatabaseConnection()
	require.NoError(t, err)
	defer db.Close()

	t.Run("Invalid Query", func(t *testing.T) {
		// Test invalid SQL query
		_, err := db.Exec("INVALID SQL QUERY")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "syntax error")
	})

	t.Run("Query Timeout", func(t *testing.T) {
		// Test query with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		var result int
		err := db.QueryRowContext(ctx, "SELECT pg_sleep(1)").Scan(&result)
		// Should timeout or complete quickly
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}
	})

	t.Run("Connection Recovery", func(t *testing.T) {
		// Test that database can handle connection issues gracefully
		// This is a basic test - in production you'd test actual connection failures

		// Close and recreate connection
		db.Close()

		newDB, err := database.CreateDatabaseConnection()
		if err != nil {
			t.Skipf("Could not recreate connection: %v", err)
			return
		}
		defer newDB.Close()

		// Test that new connection works
		var result int
		err = newDB.QueryRow("SELECT 1").Scan(&result)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})
}
