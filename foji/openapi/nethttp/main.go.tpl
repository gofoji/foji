package {{ .PackageName }}

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bir/iken/errs"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	address := "localhost:3000"

	l := setupLogging(true)

	l.Info().Msg("startup")

	r := mux.NewRouter()

    RegisterHTTP(NewServiceImpl(), r, ErrorHandler
{{- range $security, $value := .File.API.Components.SecuritySchemes -}}
	, {{ pascal $security }}Auth()
{{- end -}}
)

	l.Info().Msgf("Serving on: http://%s", address)

	if err := http.ListenAndServe(address, r); err != nil {
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
