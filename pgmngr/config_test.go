package pgmngr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_dbURL(t *testing.T) {
	t.Run("all connection information", func(t *testing.T) {
		config := Config{}
		config.Connection.Migration.Host = "localhost"
		config.Connection.Migration.Username = "test"
		config.Connection.Migration.Password = "test"
		config.Connection.Migration.Database = "test_db"
		config.Connection.Migration.QueryParams = map[string]string{
			"sslmode": "disable",
			"test":    "test",
		}

		u, err := config.dbURL()
		require.NoError(t, err)
		require.Equal(
			t,
			"postgres://test:test@localhost/test_db?sslmode=disable&test=test",
			u,
		)
	})

	t.Run("no password", func(t *testing.T) {
		config := Config{}
		config.Connection.Migration.Host = "localhost"
		config.Connection.Migration.Username = "test"
		config.Connection.Migration.Database = "test_db"
		config.Connection.Migration.QueryParams = map[string]string{
			"sslmode": "disable",
			"test":    "test",
		}

		u, err := config.dbURL()
		require.NoError(t, err)
		require.Equal(
			t,
			"postgres://test:@localhost/test_db?sslmode=disable&test=test",
			u,
		)
	})

	t.Run("no query params", func(t *testing.T) {
		config := Config{}
		config.Connection.Migration.Host = "localhost"
		config.Connection.Migration.Username = "test"
		config.Connection.Migration.Password = "test"
		config.Connection.Migration.Database = "test_db"

		u, err := config.dbURL()
		require.NoError(t, err)
		require.Equal(
			t,
			"postgres://test:test@localhost/test_db",
			u,
		)
	})
}
