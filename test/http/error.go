package http

import (
	"errors"

	"github.com/bir/iken/fastutil"
	"github.com/gofoji/foji/test"
	"github.com/valyala/fasthttp"
)

// ErrorHandler maps all error types into HTTP responses.  See fastutil.ErrorHandler for examples.
func ErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	if errors.Is(err, test.ErrNotImplemented) {
		ctx.Error("NOT IMPLEMENTED", fasthttp.StatusNotImplemented)

		return
	}

	// Add custom error handling here
	fastutil.ErrorHandler(ctx, err)
}
