package storage

import (
	"errors"
	"github.com/pingcap/failpoint"
	"github.com/rs/xid"
	"github.com/speps/go-hashids"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io/ioutil"
	"math"
	"os"
	"testing"
)

const packagePath = "auto/internal/storage/"

// setTempDir create temporary directory for database files
func setTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "db-*")
	require.NoError(t, err)

	return dir
}

// cleanUp removes temporary directory and its content
func cleanUp(t *testing.T, path string) {
	err := os.RemoveAll(path)
	require.NoError(t, err)
}

func TestNewStorageWithoutLogger(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	_, err := New(nil, dir)
	require.Equal(t, errors.New("no logger provided"), err)
}

func TestNew_ErrOpenDB(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	err := failpoint.Enable("auto/internal/storage/openDatabaseErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable("auto/internal/storage/openDatabaseErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	_, err = New(logger, dir)
	require.Equal(t, errors.New("mock open database error"), err)
}

func TestNew_ErrGetSequence(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	err := failpoint.Enable(packagePath+"getSequenceErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable(packagePath + "getSequenceErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	_, err = New(logger, dir)
	require.Equal(t, errors.New("mock get sequence error"), err)
}

func TestNew_ErrNewWithData(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	err := failpoint.Enable(packagePath+"newWithDataErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable(packagePath + "newWithDataErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	_, err = New(logger, dir)
	require.Equal(t, errors.New("mock NewWithData error"), err)
}

func TestClose(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, err := New(logger, dir)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)
}

func TestClose_ErrReleaseSequence(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	err := failpoint.Enable(packagePath+"releaseSequenceOnCloseErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable(packagePath + "releaseSequenceOnCloseErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, err := New(logger, dir)
	require.NoError(t, err)

	err = s.Close()
	require.Equal(t, errors.New("mock release sequence error"), err)
}

func TestClose_ErrCloseDatabase(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	err := failpoint.Enable(packagePath+"closeDatabaseErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable(packagePath + "closeDatabaseErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, err := New(logger, dir)
	require.NoError(t, err)

	err = s.Close()
	require.Equal(t, errors.New("mock close database error"), err)
}

func TestSaveURL_ErrNextID(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	err := failpoint.Enable(packagePath+"nextIDErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable(packagePath + "nextIDErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, err := New(logger, dir)
	require.NoError(t, err)
	defer func() {
		err = s.Close()
		require.NoError(t, err)
	}()

	_, err = s.SaveURL(0, "https://pingcap.com/blog/design-and-implementation-of-golang-failpoints")
	require.Equal(t, errors.New("mock next ID error"), err)
}

func TestSaveURL_ErrUpdate(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	err := failpoint.Enable(packagePath+"updateErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable(packagePath + "updateErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, err := New(logger, dir)
	require.NoError(t, err)
	defer func() {
		err = s.Close()
		require.NoError(t, err)
	}()

	_, err = s.SaveURL(0, "https://pingcap.com/blog/design-and-implementation-of-golang-failpoints")
	require.Equal(t, errors.New("mock update error"), err)
}

func TestGetURL_ErrInvalidShort(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, err := New(logger, dir)
	require.NoError(t, err)
	defer func() {
		err = s.Close()
		require.NoError(t, err)
	}()

	// generate short form with same length but with random salt
	// so it won't be decode by current implementation without salt
	data := hashids.NewData()
	data.MinLength = 7
	data.Salt = xid.New().String()

	hashID, err := hashids.NewWithData(data)
	require.NoError(t, err)

	short, err := hashID.EncodeInt64([]int64{0})
	require.NoError(t, err)

	_, err = s.GetURL(0, short)
	require.Equal(t, ErrInvalidShort, err)
}

func TestGetURL_ErrValueCopy(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	err := failpoint.Enable(packagePath+"valueCopyErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable(packagePath + "valueCopyErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, err := New(logger, dir)
	require.NoError(t, err)
	defer func() {
		err = s.Close()
		require.NoError(t, err)
	}()

	short, err := s.SaveURL(0, "https://www.confluent.io/blog/measure-go-code-coverage-with-bincover/")
	require.NoError(t, err)

	_, err = s.GetURL(0, short)
	require.Equal(t, errors.New("mock value copy error"), err)
}

func TestGetURL_ErrShotNotExist(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, err := New(logger, dir)
	require.NoError(t, err)
	defer func() {
		err = s.Close()
		require.NoError(t, err)
	}()

	// encoding with current storage hashID instance to ensure the short is valid for decoding
	short, err := s.hashID.EncodeInt64([]int64{0})
	require.NoError(t, err)

	_, err = s.GetURL(0, short)
	require.Equal(t, ErrShortNotExist, err)
}

func TestUint64ToInt64Slice(t *testing.T) {
	require.Equal(
		t,
		[]int64{math.MaxInt64, math.MaxInt64, 1},
		uint64ToInt64Slice(math.MaxUint64),
	)
	require.Equal(
		t,
		[]int64{math.MaxInt64, 1},
		uint64ToInt64Slice(uint64(math.MaxInt64+1)),
	)
	require.Equal(
		t,
		[]int64{1},
		uint64ToInt64Slice(uint64(1)),
	)
}

func TestSaveGetURL(t *testing.T) {
	dir := setTempDir(t)
	defer cleanUp(t, dir)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, err := New(logger, dir)
	require.NoError(t, err)
	defer func() {
		err = s.Close()
		require.NoError(t, err)
	}()

	expected := "https://pkg.go.dev/github.com/dgraph-io/badger?tab=doc#Item"

	short, err := s.SaveURL(0, expected)
	require.NoError(t, err)

	actual, err := s.GetURL(0, short)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
