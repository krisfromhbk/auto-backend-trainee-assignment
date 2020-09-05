package server

import (
	"auto/internal/storage"
	mytesting "auto/internal/testing"
	"github.com/caarlos0/env/v6"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"testing"
)

func TestWithEnvConfig(t *testing.T) {
	err := os.Setenv("HOST", "1.2.3.4")
	require.NoError(t, err)
	defer func() {
		err = os.Unsetenv("HOST")
		require.NoError(t, err)
	}()

	err = os.Setenv("PORT", "43210")
	require.NoError(t, err)
	defer func() {
		err = os.Unsetenv("HOST")
		require.NoError(t, err)
	}()

	srvCfg := EnvConfig{}
	err = env.Parse(&srvCfg)
	require.NoError(t, err)

	dir := mytesting.SetTempDir(t)
	defer mytesting.CleanUp(t, dir)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	store, err := storage.New(logger, dir)
	require.NoError(t, err)
	defer func() {
		err = store.Close()
		require.NoError(t, err)
	}()

	srv, err := New(logger, store, WithEnvConfig(srvCfg))
	require.NoError(t, err)

	require.Equal(t, "1.2.3.4:43210", srv.addr)
}
