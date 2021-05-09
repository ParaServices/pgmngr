package pgmngr

import (
	"testing"

	"github.com/icrowley/fake"
	"github.com/stretchr/testify/require"
)

func TestCreateAndDropDatabase(t *testing.T) {
	t.Run("db already exists", func(t *testing.T) {
		cfg := testConfig(t)
		cfg.Connection.Admin.Database = "template1"
		err := CreateDatabase(*cfg)
		require.Error(t, err)
	})

	t.Run("database does not exist", func(t *testing.T) {
		dbName := "pgmngr_test_" + fake.Word()
		cfg := testConfig(t)
		cfg.Connection.Migration.Database = dbName
		var err error
		err = CreateDatabase(*cfg)
		require.NoError(t, err)
		err = DropDatabase(*cfg)
		require.NoError(t, err)
		exists, err := dbExists(*cfg)
		require.NoError(t, err)
		require.False(t, exists)
	})
	t.Run("database reset", func(t *testing.T) {
		dbName := "pgmngr_test_" + fake.Word()
		cfg := testConfig(t)
		cfg.Connection.Migration.Database = dbName
		var err error
		err = CreateDatabase(*cfg)
		require.NoError(t, err)
		err = ResetDatabase(*cfg)
		require.NoError(t, err)
		exists, err := dbExists(*cfg)
		require.NoError(t, err)
		require.True(t, exists)
	})
}
