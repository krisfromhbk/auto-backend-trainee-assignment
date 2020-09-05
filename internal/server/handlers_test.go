package server

import (
	"auto/internal/storage"
	mytesting "auto/internal/testing"
	"github.com/pingcap/failpoint"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
	"net"
	"testing"
)

func serve(handler fasthttp.RequestHandler, req *fasthttp.Request, res *fasthttp.Response) error {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go func() {
		err := fasthttp.Serve(ln, handler)
		if err != nil {
			panic(err)
		}
	}()

	client := fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}

	return client.Do(req, res)
}

func TestSaveUrl_NotPOST(t *testing.T) {
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

	h := &handler{
		logger:  logger,
		Storage: store,
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.Header.SetHost("dab")
	req.SetRequestURI("/api/shorten")

	res := fasthttp.AcquireResponse()

	err = serve(h.saveURL, req, res)
	require.NoError(t, err)

	require.Equal(t, fasthttp.StatusMethodNotAllowed, res.StatusCode())
	require.Equal(t, []byte("Method Not Allowed"), res.Body())
}

func TestSaveUrl_NoUrlField(t *testing.T) {
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

	h := &handler{
		logger:  logger,
		Storage: store,
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("POST")
	req.Header.SetHost("dab")
	req.SetRequestURI("/api/shorten")
	req.AppendBody([]byte(`{"foo":"bar"}`))

	res := fasthttp.AcquireResponse()

	err = serve(h.saveURL, req, res)
	require.NoError(t, err)

	require.Equal(t, fasthttp.StatusBadRequest, res.StatusCode())
	require.Equal(t, []byte("Missing \"url\" field\n"), res.Body())
}

func TestSaveUrl_BadUrl(t *testing.T) {
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

	h := &handler{
		logger:  logger,
		Storage: store,
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("POST")
	req.Header.SetHost("dab")
	req.Header.SetContentType("application/json")
	req.SetRequestURI("/api/shorten")
	req.SetBody([]byte(`{"url":""}`))

	res := fasthttp.AcquireResponse()

	err = serve(h.saveURL, req, res)
	require.NoError(t, err)

	require.Equal(t, fasthttp.StatusBadRequest, res.StatusCode())
	require.Equal(t, []byte("Field \"url\" must be a string and have non-zero length"), res.Body())
}

func TestSaveUrl_ISE(t *testing.T) {
	dir := mytesting.SetTempDir(t)
	defer mytesting.CleanUp(t, dir)

	err := failpoint.Enable("auto/internal/storage/nextIDErr", "return(true)")
	require.NoError(t, err)
	defer func() {
		err = failpoint.Disable("auto/internal/storage/nextIDErr")
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

	h := &handler{
		logger:  logger,
		Storage: store,
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("POST")
	req.Header.SetHost("dab")
	req.Header.SetContentType("application/json")
	req.SetRequestURI("/api/shorten")
	req.SetBody([]byte(`{"url":"https://github.com/valyala/fasthttp/blob/master/server_test.go"}`))

	res := fasthttp.AcquireResponse()

	err = serve(h.saveURL, req, res)
	require.NoError(t, err)

	require.Equal(t, fasthttp.StatusInternalServerError, res.StatusCode())
	require.Equal(t, []byte("Something went wrong"), res.Body())
}

func TestGetUrl_InvalidPath(t *testing.T) {
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

	h := &handler{
		logger:  logger,
		Storage: store,
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.Header.SetHost("dab")
	req.SetRequestURI("/abcdefgh")

	res := fasthttp.AcquireResponse()

	err = serve(h.getURL, req, res)
	require.NoError(t, err)

	require.Equal(t, fasthttp.StatusBadRequest, res.StatusCode())
	require.Equal(t, []byte("Invalid path"), res.Body())
}

func TestGetUrl_InvalidShort(t *testing.T) {
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

	h := &handler{
		logger:  logger,
		Storage: store,
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.Header.SetHost("dab")
	req.SetRequestURI("/abcdefg")

	res := fasthttp.AcquireResponse()

	err = serve(h.getURL, req, res)
	require.NoError(t, err)

	require.Equal(t, fasthttp.StatusNotFound, res.StatusCode())
	require.Equal(t, []byte("404 Page not found"), res.Body())
}

func TestGetUrl_ShortNotExist(t *testing.T) {
	dirA := mytesting.SetTempDir(t)
	defer mytesting.CleanUp(t, dirA)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	storeA, err := storage.New(logger, dirA)
	require.NoError(t, err)
	defer func() {
		err = storeA.Close()
		require.NoError(t, err)
	}()

	short, err := storeA.SaveURL(0, "https://github.com/valyala/fasthttp/issues/36")
	require.NoError(t, err)

	dirB := mytesting.SetTempDir(t)
	defer mytesting.CleanUp(t, dirB)

	storeB, err := storage.New(logger, dirB)
	require.NoError(t, err)
	defer func() {
		err = storeB.Close()
		require.NoError(t, err)
	}()

	h := &handler{
		logger:  logger,
		Storage: storeB,
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.Header.SetHost("dab")
	req.SetRequestURI("/" + short)

	res := fasthttp.AcquireResponse()

	err = serve(h.getURL, req, res)
	require.NoError(t, err)

	require.Equal(t, fasthttp.StatusNotFound, res.StatusCode())
	require.Equal(t, []byte("404 Page not found"), res.Body())
}

func TestGetUrl_ISE(t *testing.T) {
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
	defer func() {
		err = store.Close()
		require.NoError(t, err)
	}()

	short, err := store.SaveURL(0, "https://godoc.org/github.com/valyala/fasthttp")
	require.NoError(t, err)

	h := &handler{
		logger:  logger,
		Storage: store,
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.Header.SetHost("dab")
	req.SetRequestURI("/" + short)

	res := fasthttp.AcquireResponse()

	err = serve(h.getURL, req, res)
	require.NoError(t, err)

	require.Equal(t, fasthttp.StatusInternalServerError, res.StatusCode())
	require.Equal(t, []byte("Something went wrong"), res.Body())
}

func TestSaveGetUrl(t *testing.T) {
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

	h := &handler{
		logger:  logger,
		Storage: store,
	}

	saveReq := fasthttp.AcquireRequest()
	saveReq.Header.SetMethod("POST")
	saveReq.Header.SetHost("dab")
	saveReq.Header.SetContentType("application/json")
	saveReq.SetRequestURI("/api/shorten")
	saveReq.SetBody([]byte(`{"url":"https://github.com/valyala/fasthttp"}`))

	saveRes := fasthttp.AcquireResponse()
	err = serve(h.saveURL, saveReq, saveRes)
	require.NoError(t, err)
	short := fastjson.GetString(saveRes.Body(), "short")

	getReq := fasthttp.AcquireRequest()
	getReq.Header.SetMethod("GET")
	getReq.Header.SetHost("dab")
	getReq.SetRequestURI("/" + short)

	getRes := fasthttp.AcquireResponse()
	err = serve(h.getURL, getReq, getRes)
	require.NoError(t, err)

	require.Equal(t, fasthttp.StatusMovedPermanently, getRes.StatusCode())
	require.Equal(t, []byte("https://github.com/valyala/fasthttp"), getRes.Header.Peek("Location"))
}
