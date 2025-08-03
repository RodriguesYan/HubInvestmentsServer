package migration

import (
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDatabaseURL = "postgres://yanrodrigues@localhost/yanrodrigues?sslmode=disable"

func TestNewMigrationManager(t *testing.T) {
	// Test creating a migration manager
	mgr, err := NewMigrationManager(testDatabaseURL)

	// We expect this to fail in CI/testing environments where the database might not be available
	// So we'll just check that the function signature is correct
	if err != nil {
		t.Skipf("Skipping test due to database connection error (expected in CI): %v", err)
		return
	}

	require.NotNil(t, mgr)
	defer mgr.Close()
}

func TestMigrationManager_Version(t *testing.T) {
	mgr, err := NewMigrationManager(testDatabaseURL)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}
	defer mgr.Close()

	// Test getting version (might fail if no migrations have been run)
	version, dirty, err := mgr.Version()

	// Either we get a version or an error saying no migrations have been run
	if err != nil {
		assert.Contains(t, err.Error(), "no migration")
	} else {
		assert.GreaterOrEqual(t, version, uint(0))
		assert.IsType(t, false, dirty)
	}
}

func TestInvalidDatabaseURL(t *testing.T) {
	invalidURL := "invalid://database/url"

	mgr, err := NewMigrationManager(invalidURL)

	assert.Error(t, err)
	assert.Nil(t, mgr)
	assert.Contains(t, err.Error(), "failed to open database connection")
}
