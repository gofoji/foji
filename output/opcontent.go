package output

import "github.com/getkin/kin-openapi/openapi3"

const (
	ApplicationJSON  = "application/json"
	ApplicationJSONL = "application/jsonl"
	TextPlain        = "text/plain"
	TextHTML         = "text/html"
)

type OpResponse struct {
	Key       string
	MimeType  string
	MediaType *openapi3.MediaType
	GoType    string
}

func (o OpResponse) IsJson() bool {
	return o.MimeType == ApplicationJSON
}

func (o OpResponse) IsText() bool {
	return o.MimeType == TextPlain
}

func (o OpResponse) IsHTML() bool {
	return o.MimeType == TextHTML
}

func (o OpResponse) IsLongPollingOperation() bool {
	return o.MimeType == ApplicationJSONL
}

type OpBody struct {
	MimeType string
	Schema   *openapi3.SchemaRef
}

func (o OpBody) IsJson() bool {
	return o.MimeType == ApplicationJSON
}

func (o OpBody) IsText() bool {
	return o.MimeType == TextPlain
}

func (o OpBody) IsHTML() bool {
	return o.MimeType == TextHTML
}
