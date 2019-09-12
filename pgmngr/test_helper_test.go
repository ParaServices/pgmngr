package pgmngr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testCliContext struct{}

func (t *testCliContext) String(s string) string {
	return configFile
}

func testConfig(t *testing.T) *Config {
	cfg := &Config{}
	err := LoadConfig(&testCliContext{}, cfg)
	require.NoError(t, err)
	return cfg
}
