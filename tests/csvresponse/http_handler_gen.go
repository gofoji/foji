// Code generated by foji (dev build), template: foji/openapi/handler.go.tpl; DO NOT EDIT.

package csvresponse

import (
	"context"
	"io"
	"net/http"

	"github.com/bir/iken/httputil"
	"github.com/bir/iken/logctx"
)

type Operations interface {
	GetByteCsv(ctx context.Context) ([]byte, error)
	GetReaderCsv(ctx context.Context) (io.Reader, error)
	GetStringCsv(ctx context.Context) (string, error)
}

type OpenAPIHandlers struct {
	ops Operations
}

type Mux interface {
	Handle(pattern string, handler http.Handler)
}

func RegisterHTTP(ops Operations, r Mux) *OpenAPIHandlers {
	s := OpenAPIHandlers{
		ops: ops,
	}

	r.Handle("GET /bytesCSV", http.HandlerFunc(s.GetByteCsv))
	r.Handle("GET /readerCSV", http.HandlerFunc(s.GetReaderCsv))
	r.Handle("GET /stringCSV", http.HandlerFunc(s.GetStringCsv))

	return &s
}

// GetByteCsv
func (h OpenAPIHandlers) GetByteCsv(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getByteCSV")

	response, err := h.ops.GetByteCsv(r.Context())
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.Write(w, r, "text/csv", 200, response)
}

// GetReaderCsv
func (h OpenAPIHandlers) GetReaderCsv(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getReaderCSV")

	response, err := h.ops.GetReaderCsv(r.Context())
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.ReaderWrite(w, r, "text/csv", 200, response)
}

// GetStringCsv
func (h OpenAPIHandlers) GetStringCsv(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getStringCSV")

	response, err := h.ops.GetStringCsv(r.Context())
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.Write(w, r, "text/csv", 200, []byte(response))
}
