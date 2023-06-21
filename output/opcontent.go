package output

import "github.com/getkin/kin-openapi/openapi3"

const (
	ApplicationJSON  = "application/json"
	ApplicationJSONL = "application/jsonl"
	TextPlain        = "text/plain"
	TextHTML         = "text/html"
	TextCSV          = "text/csv"
)

type MimeType string

func (m MimeType) IsJson() bool {
	return m == ApplicationJSON
}

func (m MimeType) IsText() bool {
	return m == TextPlain
}

func (m MimeType) IsHTML() bool {
	return m == TextHTML
}
func (m MimeType) IsCSV() bool {
	return m == TextCSV
}

func (m MimeType) IsLongPollingOperation() bool {
	return m == ApplicationJSONL
}

func (m MimeType) String() string { return string(m) }

type OpResponse struct {
	Key string
	MimeType
	MediaType *openapi3.MediaType
	GoType    string
}

type OpBody struct {
	MimeType
	Schema *openapi3.SchemaRef
}
