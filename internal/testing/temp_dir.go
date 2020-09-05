package testing

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

// SetTempDir create temporary directory for database files
func SetTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "db-*")
	require.NoError(t, err)

	return dir
}

// CleanUp removes temporary directory and its content
func CleanUp(t *testing.T, path string) {
	err := os.RemoveAll(path)
	require.NoError(t, err)
}
