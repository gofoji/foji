package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/bir/iken/chain"
	"github.com/bir/iken/errs"
	"github.com/bir/iken/fastutil"
	"github.com/bir/iken/notify"
	"github.com/fasthttp/router"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
"{{ .Params.Package }}"
"{{ .Params.Package }}/http"
)

func main() {
	address := "localhost:3000"

	r := router.New()
	l := setupLogging(true)

	n := notify.NewZerolog(l)
	_, _ = n.Send("startup")

	defer notify.Monitor(n)

	r.PanicHandler = fastutil.PanicHandler
    http.RegisterHTTP(test.New(), r, http.ErrorHandler
{{- range $security, $value := .File.API.Components.SecuritySchemes -}}
	, {{ $.PackageName }}.{{ pascal $security }}Auth()
{{- end -}}
)

	server := &fasthttp.Server{}
	server.NoDefaultServerHeader = true
	c := chain.New(fasthttp.CompressHandler,
		fastutil.RequestLogger(l, n, false, true, true))

	server.Handler = c.Handler(r.Handler)

	l.Info().Msgf("Serving on: http://%s", address)

	if err := server.ListenAndServe(address); err != nil {
		log.Err(err)
	}
}

func setupLogging(consoleLog bool) zerolog.Logger {
	zerolog.DurationFieldInteger = true
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.ErrorStackMarshaler = errs.MarshalStack

	var out io.Writer = os.Stdout

	if consoleLog {
		out = zerolog.NewConsoleWriter()
	}

	return log.Output(out)
}
