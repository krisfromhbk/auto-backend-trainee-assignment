package server

import (
	"auto/internal/storage"
	"errors"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
	"strings"
)

type handler struct {
	logger  *zap.Logger
	Storage *storage.Storage
}

// saveURL handles HTTP requests on "/api/shorten" endpoint
func (h *handler) saveURL(ctx *fasthttp.RequestCtx) {
	logger := h.logger.With(zap.Uint64("request id", ctx.ID()), zap.String("path", string(ctx.Path())))
	logger.Debug("New request")

	if !ctx.IsPost() {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		ctx.SetBody([]byte("Method Not Allowed"))
		return
	}

	if !fastjson.Exists(ctx.PostBody(), "url") {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte("Missing \"url\" field\n"))
		return
	}

	url := fastjson.GetString(ctx.PostBody(), "url")
	if len(url) == 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte("Field \"url\" must be a string and have non-zero length"))
		return
	}

	short, err := h.Storage.SaveURL(ctx.ID(), url)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte("Something went wrong"))
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody([]byte(`{"short":"` + short + `"}`))

	logger.Debug("Finishing request")

	return
}

// getURL handles HTTP requests on "/**" endpoint. Returns corresponding redirect or NotFound
func (h *handler) getURL(ctx *fasthttp.RequestCtx) {
	logger := h.logger.With(zap.Uint64("request id", ctx.ID()), zap.String("path", string(ctx.Path())))
	logger.Debug("New request")

	path := strings.Trim(string(ctx.Path()), "/")
	if len(path) != 7 {
		ctx.Error("Invalid path", fasthttp.StatusBadRequest)
		return
	}

	url, err := h.Storage.GetURL(ctx.ID(), path)
	if err != nil {
		if errors.Is(err, storage.ErrShortNotExist) || errors.Is(err, storage.ErrInvalidShort) {
			ctx.NotFound()
			return
		}

		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte("Something went wrong"))
		return
	}

	ctx.Redirect(url, fasthttp.StatusMovedPermanently)

	logger.Debug("Finishing request")

	return
}
