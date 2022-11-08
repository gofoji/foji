package output

import "github.com/getkin/kin-openapi/openapi3"

type OpResponse struct {
	Key       string
	MimeType  string
	MediaType *openapi3.MediaType
	GoType    string
}

func (op OpResponse) IsLongPollingOperation() bool {
	return op.MimeType == "application/jsonl"
}
