// Package errs provides unified error types and constants used throughout foji.
package errs

// Error is a simple error string type used for custom error definitions.
type Error string

// Error implements the error interface for Error type.
func (e Error) Error() string {
	return string(e)
}

const (
	// ErrRuntime indicates an error occurred in template runtime functions.
	ErrRuntime = Error("runtime")

	// ErrWeld indicates an error occurred during the welding process.
	ErrWeld = Error("welding error")

	// ErrMissingRequirement indicates a required condition was not met.
	ErrMissingRequirement = Error("requires")

	// ErrNotNeeded indicates the output generation should be skipped.
	ErrNotNeeded = Error("not needed")

	// ErrPermExists indicates a permanent file (prefixed with !) already exists.
	ErrPermExists = Error("file exists")

	// ErrInvalidDictParams indicates invalid parameters in WithParams call.
	ErrInvalidDictParams = Error("invalid dict params in call to WithParams, must be key and value pairs")

	// ErrInvalidDictKey indicates an invalid dictionary key in WithParams call.
	ErrInvalidDictKey = Error("invalid dict params in call to WithParams, must be key and value pairs")
)
