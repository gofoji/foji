package http

import (
    "errors"

    "github.com/bir/iken/fastutil"
    "github.com/valyala/fasthttp"
    "{{ .Params.Package }}"
)

// ErrorHandler maps all error types into HTTP responses.  See fastutil.ErrorHandler for examples.
func ErrorHandler(ctx *fasthttp.RequestCtx, err error) {
    if errors.Is(err, {{ .PackageName }}.ErrNotImplemented) {
        ctx.Error("NOT IMPLEMENTED", fasthttp.StatusNotImplemented)

        return
    }

    // Add custom error handling here
    fastutil.ErrorHandler(ctx, err)
}
