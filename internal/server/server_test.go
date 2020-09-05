package server

import (
	"auto/internal/storage"
	mytesting "auto/internal/testing"
	"bufio"
	"errors"
	"fmt"
	"github.com/pingcap/failpoint"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"go.uber.org/zap"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestNew_NoLogger(t *testing.T) {
	_, err := New(nil, nil)
	require.Equal(t, errors.New("no logger provided"), err)
}

func TestNew_NoStorage(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	_, err = New(logger, nil)
	require.Equal(t, errors.New("no storage provided"), err)
}

func TestServerSwitch(t *testing.T) {
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

	srv, err := New(logger, store)
	require.NoError(t, err)

	ln := fasthttputil.NewInmemoryListener()

	serverCh := make(chan struct{})
	go func() {
		err := srv.httpServer.Serve(ln)
		require.NoError(t, err)
		close(serverCh)
	}()

	clientCh := make(chan struct{})
	go func() {
		saveUrlConn, err := ln.Dial()
		require.NoError(t, err)

		_, err = saveUrlConn.Write([]byte("GET /api/shorten HTTP/1.1\r\nHost: aa\r\n\r\n"))
		require.NoError(t, err)

		saveReader := bufio.NewReader(saveUrlConn)
		var saveRes fasthttp.Response
		err = saveRes.Read(saveReader)

		require.Equal(t, fasthttp.StatusMethodNotAllowed, saveRes.StatusCode())

		err = saveUrlConn.Close()
		require.NoError(t, err)

		getUrlConn, err := ln.Dial()
		require.NoError(t, err)

		_, err = getUrlConn.Write([]byte("GET /abcdefg HTTP/1.1\r\nHost: aa\r\n\r\n"))
		require.NoError(t, err)

		getReader := bufio.NewReader(getUrlConn)
		var getRes fasthttp.Response
		err = getRes.Read(getReader)

		require.Equal(t, fasthttp.StatusNotFound, getRes.StatusCode())

		err = getUrlConn.Close()
		require.NoError(t, err)

		close(clientCh)
	}()

	select {
	case <-clientCh:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	err = ln.Close()
	require.NoError(t, err)

	select {
	case <-serverCh:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestStart(t *testing.T) {
	dir := mytesting.SetTempDir(t)
	defer mytesting.CleanUp(t, dir)

	err := failpoint.Enable("auto/internal/storage/valueCopyErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable("auto/internal/storage/valueCopyErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	store, err := storage.New(logger, dir)
	require.NoError(t, err)

	srv, err := New(logger, store)
	require.NoError(t, err)

	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	var startErr error
	go func() {
		startErr = srv.Start()
	}()

	time.Sleep(1 * time.Second)

	err = p.Signal(syscall.SIGINT)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	require.NoError(t, startErr)
}

func TestStart_ErrListenAndServe(t *testing.T) {
	dir := mytesting.SetTempDir(t)
	defer mytesting.CleanUp(t, dir)

	err := failpoint.Enable("auto/internal/server/listenAndServeErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable("auto/internal/server/listenAndServeErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	store, err := storage.New(logger, dir)
	require.NoError(t, err)
	defer func() {
		err = store.Close()
		require.NoError(t, err)
	}()

	srv, err := New(logger, store)
	require.NoError(t, err)

	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	var startErr error
	go func() {
		startErr = srv.Start()
	}()

	time.Sleep(1 * time.Second)

	err = p.Signal(syscall.SIGINT)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	expected := fmt.Errorf("ListenAndServe error: %w", errors.New("mock listen and serve error"))
	require.Equal(t, expected, startErr)
}

func TestStart_ErrShutdown(t *testing.T) {
	dir := mytesting.SetTempDir(t)
	defer mytesting.CleanUp(t, dir)

	err := failpoint.Enable("auto/internal/server/shutdownErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable("auto/internal/server/shutdownErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	store, err := storage.New(logger, dir)
	require.NoError(t, err)

	srv, err := New(logger, store)
	require.NoError(t, err)

	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	var startErr error
	go func() {
		startErr = srv.Start()
	}()

	time.Sleep(1 * time.Second)

	err = p.Signal(syscall.SIGINT)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	require.NoError(t, startErr)
}

func TestStart_ErrAfterShutdown(t *testing.T) {
	dir := mytesting.SetTempDir(t)
	defer mytesting.CleanUp(t, dir)

	err := failpoint.Enable("auto/internal/storage/releaseSequenceOnCloseErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable("auto/internal/storage/releaseSequenceOnCloseErr")
		require.NoError(t, err)
	}()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	store, err := storage.New(logger, dir)
	require.NoError(t, err)

	srv, err := New(logger, store)
	require.NoError(t, err)

	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	var startErr error
	go func() {
		startErr = srv.Start()
	}()

	time.Sleep(1 * time.Second)

	err = p.Signal(syscall.SIGINT)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	require.Equal(t, errors.New("mock release sequence error"), startErr)
}
