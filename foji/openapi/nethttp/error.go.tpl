package {{ .PackageName }}

import (
    "errors"
    "net/http"

    "github.com/bir/iken/httputil"
)

type ServiceError string

func (e ServiceError) Error() string {
	return string(e)
}

const ErrNotImplemented = ServiceError("not implemented")


// ErrorHandler maps all error types into HTTP responses.  See fastutil.ErrorHandler for examples.
func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	switch e := err.(type) { //nolint:errorlint // false positive
	case ServiceError:
		if errors.Is(e, ErrNotImplemented) {
			if err := httputil.JSONWrite(w, http.StatusNotImplemented, e); err != nil {
				panic(err)
			}

			return
		}
	}

    // Add custom error handling here
    httputil.ErrorHandler(w, r, err)
}
