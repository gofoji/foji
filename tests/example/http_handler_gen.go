// Code generated by foji (dev build), template: foji/openapi/handler.go.tpl; DO NOT EDIT.

package example

import (
	"context"
	"net/http"
	"time"

	"github.com/bir/iken/httputil"
	"github.com/bir/iken/logctx"
	"github.com/bir/iken/params"
	"github.com/bir/iken/validation"
	"github.com/google/uuid"
)

type (
	RequestAuthenticator = httputil.AuthenticateFunc[*ExampleAuth]
	TokenAuthenticator   = httputil.TokenAuthenticatorFunc[*ExampleAuth]

	SecurityGroup  = httputil.SecurityGroup[*ExampleAuth]
	SecurityGroups = httputil.SecurityGroups[*ExampleAuth]
	AuthorizeFunc  = httputil.AuthorizeFunc[*ExampleAuth]
)

type Operations interface {
	GetExamples(ctx context.Context) (*Examples, error)
	GetAuthComplex(ctx context.Context, user *ExampleAuth) error
	GetAuthSimple(ctx context.Context, user *ExampleAuth) error
	GetAuthSimpleMaybe(ctx context.Context, user *ExampleAuth) error
	GetAuthSimple2(ctx context.Context, user *ExampleAuth) error
	GetAuthSimple2Maybe(ctx context.Context, user *ExampleAuth) error
	GetAuthComplexMaybe(ctx context.Context, user *ExampleAuth) error
	GetComplexSecurity(ctx context.Context, user *ExampleAuth) error
	AddForm(ctx context.Context, body AddFormRequest) (*FooBar, error)
	AddMultipartForm(ctx context.Context, body AddMultipartFormRequest) (*FooBar, error)
	HeaderResponse(ctx context.Context) (http.Header, error)
	AddInlinedAllOf(ctx context.Context, body AddInlinedAllOfRequest) (*FooBar, error)
	AddInlinedBody(ctx context.Context, body AddInlinedBodyRequest) (*FooBar, error)
	GetExampleParams(ctx context.Context, k1 string, k2 uuid.UUID, k3 time.Time, k4 int32, k5 int64, enumTest GetExampleParamsEnumTest) (*Example, error)
	NoResponse(ctx context.Context, body Foo) error
	GetExampleOptional(ctx context.Context, k1 *string, k2 *uuid.UUID, k3 *time.Time, k4 *int32, k5 *int64, k5Default int64) (*Example, error)
	GetExampleQuery(ctx context.Context, k1 string, k2 uuid.UUID, k3 time.Time, k4 int32, k5 int64) (*Example, error)
	GetRawRequest(r *http.Request, vehicle GetRawRequestVehicle) (*Example, error)
	GetRawRequestResponse(r *http.Request, w http.ResponseWriter, vehicle GetRawRequestResponseVehicle) (*Example, error)
	GetRawRequestResponseAndHeaders(r *http.Request, w http.ResponseWriter, vehicle GetRawRequestResponseAndHeadersVehicle) (*Example, http.Header, error)
	GetRawResponse(ctx context.Context, w http.ResponseWriter, vehicle GetRawResponseVehicle) (*Example, error)
	GetTest(ctx context.Context, vehicle GetTestVehicle, vehicleDefault GetTestVehicleDefault, playerID uuid.UUID, color ColorQuery, colorDefault ColorQueryDefault, season Season) (*Example, error)
}

type OpenAPIHandlers struct {
	ops                         Operations
	headerAuthAuth              RequestAuthenticator
	jwtAuth                     RequestAuthenticator
	rawAuth                     RequestAuthenticator
	authorize                   AuthorizeFunc
	getAuthComplexSecurity      SecurityGroups
	getAuthSimpleMaybeSecurity  SecurityGroups
	getAuthSimple2MaybeSecurity SecurityGroups
	getAuthComplexMaybeSecurity SecurityGroups
}

type Mux interface {
	Handle(pattern string, handler http.Handler)
}

func RegisterHTTP(ops Operations, r Mux, headerAuthAuth TokenAuthenticator, jwtAuth TokenAuthenticator, rawAuth RequestAuthenticator, authorize AuthorizeFunc) *OpenAPIHandlers {
	s := OpenAPIHandlers{ops: ops, headerAuthAuth: httputil.HeaderAuth("Authorization", headerAuthAuth), jwtAuth: httputil.QueryAuth("jwt", jwtAuth), rawAuth: rawAuth, authorize: authorize}

	s.getAuthComplexSecurity = SecurityGroups{
		SecurityGroup{httputil.NewAuthCheck(s.headerAuthAuth, authorize, "foo")},
		SecurityGroup{httputil.NewAuthCheck(s.headerAuthAuth, authorize, "bar")},
		SecurityGroup{httputil.NewAuthCheck(s.jwtAuth, nil)},
	}

	s.getAuthSimpleMaybeSecurity = SecurityGroups{
		SecurityGroup{httputil.NewAuthCheck(s.headerAuthAuth, nil)},
		SecurityGroup{},
	}

	s.getAuthSimple2MaybeSecurity = SecurityGroups{
		SecurityGroup{httputil.NewAuthCheck(s.headerAuthAuth, authorize, "foo")},
		SecurityGroup{httputil.NewAuthCheck(s.headerAuthAuth, authorize, "bar")},
		SecurityGroup{},
	}

	s.getAuthComplexMaybeSecurity = SecurityGroups{
		SecurityGroup{httputil.NewAuthCheck(s.headerAuthAuth, nil)},
		SecurityGroup{httputil.NewAuthCheck(s.jwtAuth, nil)},
		SecurityGroup{},
	}

	r.Handle("GET /examples", http.HandlerFunc(s.GetExamples))
	r.Handle("GET /examples/auth/complex", http.HandlerFunc(s.GetAuthComplex))
	r.Handle("GET /examples/auth/simple", http.HandlerFunc(s.GetAuthSimple))
	r.Handle("GET /examples/auth/simple/maybe", http.HandlerFunc(s.GetAuthSimpleMaybe))
	r.Handle("GET /examples/auth/simple2", http.HandlerFunc(s.GetAuthSimple2))
	r.Handle("GET /examples/auth/simple2/maybe", http.HandlerFunc(s.GetAuthSimple2Maybe))
	r.Handle("GET /examples/complexAuthMaybe", http.HandlerFunc(s.GetAuthComplexMaybe))
	r.Handle("GET /examples/complexSecurity", http.HandlerFunc(s.GetComplexSecurity))
	r.Handle("POST /examples/form", http.HandlerFunc(s.AddForm))
	r.Handle("POST /examples/form:multipart", http.HandlerFunc(s.AddMultipartForm))
	r.Handle("GET /examples/header", http.HandlerFunc(s.HeaderResponse))
	r.Handle("POST /examples/inlinedAllOf", http.HandlerFunc(s.AddInlinedAllOf))
	r.Handle("POST /examples/inlinedBody", http.HandlerFunc(s.AddInlinedBody))
	r.Handle("GET /examples/key1/{k1}/key2/{k2}/key3/{k3}/key4/{key4}/key5/{key5}", http.HandlerFunc(s.GetExampleParams))
	r.Handle("POST /examples/noResponse", http.HandlerFunc(s.NoResponse))
	r.Handle("GET /examples/optional", http.HandlerFunc(s.GetExampleOptional))
	r.Handle("GET /examples/query", http.HandlerFunc(s.GetExampleQuery))
	r.Handle("GET /examples/rawRequest", http.HandlerFunc(s.GetRawRequest))
	r.Handle("GET /examples/rawRequestResponse", http.HandlerFunc(s.GetRawRequestResponse))
	r.Handle("GET /examples/rawRequestResponseAndHeaders", http.HandlerFunc(s.GetRawRequestResponseAndHeaders))
	r.Handle("GET /examples/rawResponse", http.HandlerFunc(s.GetRawResponse))
	r.Handle("GET /examples/test", http.HandlerFunc(s.GetTest))

	return &s
}

// GetExamples
func (h OpenAPIHandlers) GetExamples(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getExamples")

	response, err := h.ops.GetExamples(r.Context())
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// GetAuthComplex
func (h OpenAPIHandlers) GetAuthComplex(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getAuthComplex")

	user, err := h.getAuthComplexSecurity.Auth(r)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	err = h.ops.GetAuthComplex(r.Context(), user)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	w.WriteHeader(200)
}

// GetAuthSimple
func (h OpenAPIHandlers) GetAuthSimple(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getAuthSimple")

	user, err := h.headerAuthAuth(r)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	err = h.ops.GetAuthSimple(r.Context(), user)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	w.WriteHeader(200)
}

// GetAuthSimpleMaybe
func (h OpenAPIHandlers) GetAuthSimpleMaybe(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getAuthSimpleMaybe")

	user, err := h.getAuthSimpleMaybeSecurity.Auth(r)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	err = h.ops.GetAuthSimpleMaybe(r.Context(), user)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	w.WriteHeader(200)
}

// GetAuthSimple2
func (h OpenAPIHandlers) GetAuthSimple2(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getAuthSimple2")

	user, err := h.headerAuthAuth(r)
	if err == nil {
		err = h.authorize(r.Context(), user, []string{"foo"})
		if err != nil {
			err = h.authorize(r.Context(), user, []string{"bar"})
		}
	}

	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	err = h.ops.GetAuthSimple2(r.Context(), user)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	w.WriteHeader(200)
}

// GetAuthSimple2Maybe
func (h OpenAPIHandlers) GetAuthSimple2Maybe(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getAuthSimple2Maybe")

	user, err := h.getAuthSimple2MaybeSecurity.Auth(r)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	err = h.ops.GetAuthSimple2Maybe(r.Context(), user)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	w.WriteHeader(200)
}

// GetAuthComplexMaybe
func (h OpenAPIHandlers) GetAuthComplexMaybe(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getAuthComplexMaybe")

	user, err := h.getAuthComplexMaybeSecurity.Auth(r)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	err = h.ops.GetAuthComplexMaybe(r.Context(), user)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	w.WriteHeader(200)
}

// GetComplexSecurity
func (h OpenAPIHandlers) GetComplexSecurity(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getComplexSecurity")

	user, err := h.rawAuth(r)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	err = h.ops.GetComplexSecurity(r.Context(), user)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	w.WriteHeader(200)
}

// AddForm
func (h OpenAPIHandlers) AddForm(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "AddForm")

	body, err := ParseFormAddFormRequest(r)
	if err != nil {
		httputil.ErrorHandler(w, r, validation.Error{Message: "unable to parse form", Source: err})

		return
	}

	response, err := h.ops.AddForm(r.Context(), body)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// AddMultipartForm
func (h OpenAPIHandlers) AddMultipartForm(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "AddMultipartForm")

	body, err := ParseFormAddMultipartFormRequest(r)
	if err != nil {
		httputil.ErrorHandler(w, r, validation.Error{Message: "unable to parse form", Source: err})

		return
	}

	response, err := h.ops.AddMultipartForm(r.Context(), body)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// HeaderResponse
// Check header responses
func (h OpenAPIHandlers) HeaderResponse(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "HeaderResponse")

	headers, err := h.ops.HeaderResponse(r.Context())
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	location := headers.Values("Location")
	for _, v := range location {
		if v != "" {
			w.Header().Add("Location", v)
		}
	}

	w.WriteHeader(200)
}

// AddInlinedAllOf
func (h OpenAPIHandlers) AddInlinedAllOf(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "AddInlinedAllOf")

	var body AddInlinedAllOfRequest
	if err = httputil.GetJSONBody(r.Body, &body); err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	response, err := h.ops.AddInlinedAllOf(r.Context(), body)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// AddInlinedBody
func (h OpenAPIHandlers) AddInlinedBody(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "AddInlinedBody")

	var body AddInlinedBodyRequest
	if err = httputil.GetJSONBody(r.Body, &body); err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	response, err := h.ops.AddInlinedBody(r.Context(), body)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// GetExampleParams
func (h OpenAPIHandlers) GetExampleParams(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getExampleParams")

	var validationErrors validation.Errors

	k1, _, err := params.GetStringPath(r, "k1", true)
	if err != nil {
		validationErrors.Add("k1", err)
	}

	k2, _, err := params.GetUUIDPath(r, "k2", true)
	if err != nil {
		validationErrors.Add("k2", err)
	}

	k3, _, err := params.GetTimePath(r, "k3", true)
	if err != nil {
		validationErrors.Add("k3", err)
	}

	k4, _, err := params.GetInt32Path(r, "k4", true)
	if err != nil {
		validationErrors.Add("k4", err)
	}

	k5, _, err := params.GetInt64Path(r, "k5", true)
	if err != nil {
		validationErrors.Add("k5", err)
	}

	// Enum Description
	enumTest, ok, err := params.GetEnumQuery(r, "enumTest", false, NewGetExampleParamsEnumTest)

	if err != nil {
		validationErrors.Add("enumTest", err)
	} else if !ok {
		enumTest = GetExampleParamsEnumTestValueA
	}

	if validationErrors != nil {
		httputil.ErrorHandler(w, r, validationErrors.GetErr())

		return
	}

	response, err := h.ops.GetExampleParams(r.Context(), k1, k2, k3, k4, k5, enumTest)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// NoResponse
func (h OpenAPIHandlers) NoResponse(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "noResponse")

	var body Foo
	if err = httputil.GetJSONBody(r.Body, &body); err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	err = h.ops.NoResponse(r.Context(), body)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	w.WriteHeader(201)
}

// GetExampleOptional
func (h OpenAPIHandlers) GetExampleOptional(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getExampleOptional")

	var validationErrors validation.Errors

	var k1 *string

	k1Val, ok, err := params.GetStringQuery(r, "k1", false)
	if err != nil {
		validationErrors.Add("k1", err)
	}

	if ok {
		k1 = &k1Val
	}

	var k2 *uuid.UUID

	k2Val, ok, err := params.GetUUIDQuery(r, "k2", false)
	if err != nil {
		validationErrors.Add("k2", err)
	}

	if ok {
		k2 = &k2Val
	}

	var k3 *time.Time

	k3Val, ok, err := params.GetTimeQuery(r, "k3", false)
	if err != nil {
		validationErrors.Add("k3", err)
	}

	if ok {
		k3 = &k3Val
	}

	var k4 *int32

	k4Val, ok, err := params.GetInt32Query(r, "k4", false)
	if err != nil {
		validationErrors.Add("k4", err)
	}

	if ok {
		k4 = &k4Val
	}

	var k5 *int64

	k5Val, ok, err := params.GetInt64Query(r, "k5", false)
	if err != nil {
		validationErrors.Add("k5", err)
	}

	if ok {
		k5 = &k5Val
	}

	k5Default, ok, err := params.GetInt64Query(r, "k5Default", false)
	if err != nil {
		validationErrors.Add("k5Default", err)
	} else if !ok {
		k5Default = 1
	}

	if validationErrors != nil {
		httputil.ErrorHandler(w, r, validationErrors.GetErr())

		return
	}

	response, err := h.ops.GetExampleOptional(r.Context(), k1, k2, k3, k4, k5, k5Default)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// GetExampleQuery
func (h OpenAPIHandlers) GetExampleQuery(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getExampleQuery")

	var validationErrors validation.Errors

	k1, _, err := params.GetStringQuery(r, "k1", true)
	if err != nil {
		validationErrors.Add("k1", err)
	}

	k2, _, err := params.GetUUIDQuery(r, "k2", true)
	if err != nil {
		validationErrors.Add("k2", err)
	}

	k3, _, err := params.GetTimeQuery(r, "k3", true)
	if err != nil {
		validationErrors.Add("k3", err)
	}

	k4, _, err := params.GetInt32Query(r, "k4", true)
	if err != nil {
		validationErrors.Add("k4", err)
	}

	k5, _, err := params.GetInt64Query(r, "k5", true)
	if err != nil {
		validationErrors.Add("k5", err)
	}

	if validationErrors != nil {
		httputil.ErrorHandler(w, r, validationErrors.GetErr())

		return
	}

	response, err := h.ops.GetExampleQuery(r.Context(), k1, k2, k3, k4, k5)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// GetRawRequest
func (h OpenAPIHandlers) GetRawRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getRawRequest")

	var validationErrors validation.Errors

	vehicle, _, err := params.GetEnumQuery(r, "vehicle", true, NewGetRawRequestVehicle)
	if err != nil {
		validationErrors.Add("vehicle", err)
	}

	if validationErrors != nil {
		httputil.ErrorHandler(w, r, validationErrors.GetErr())

		return
	}

	response, err := h.ops.GetRawRequest(r, vehicle)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// GetRawRequestResponse
func (h OpenAPIHandlers) GetRawRequestResponse(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getRawRequestResponse")

	var validationErrors validation.Errors

	vehicle, _, err := params.GetEnumQuery(r, "vehicle", true, NewGetRawRequestResponseVehicle)
	if err != nil {
		validationErrors.Add("vehicle", err)
	}

	if validationErrors != nil {
		httputil.ErrorHandler(w, r, validationErrors.GetErr())

		return
	}

	ww := httputil.WrapWriter(w)
	w = ww

	response, err := h.ops.GetRawRequestResponse(r, w, vehicle)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	if ww.Status() > 0 {
		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// GetRawRequestResponseAndHeaders
func (h OpenAPIHandlers) GetRawRequestResponseAndHeaders(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getRawRequestResponseAndHeaders")

	var validationErrors validation.Errors

	vehicle, _, err := params.GetEnumQuery(r, "vehicle", true, NewGetRawRequestResponseAndHeadersVehicle)
	if err != nil {
		validationErrors.Add("vehicle", err)
	}

	if validationErrors != nil {
		httputil.ErrorHandler(w, r, validationErrors.GetErr())

		return
	}

	ww := httputil.WrapWriter(w)
	w = ww

	response, headers, err := h.ops.GetRawRequestResponseAndHeaders(r, w, vehicle)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	location := headers.Values("Location")
	for _, v := range location {
		if v != "" {
			w.Header().Add("Location", v)
		}
	}

	if ww.Status() > 0 {
		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// GetRawResponse
func (h OpenAPIHandlers) GetRawResponse(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getRawResponse")

	var validationErrors validation.Errors

	vehicle, _, err := params.GetEnumQuery(r, "vehicle", true, NewGetRawResponseVehicle)
	if err != nil {
		validationErrors.Add("vehicle", err)
	}

	if validationErrors != nil {
		httputil.ErrorHandler(w, r, validationErrors.GetErr())

		return
	}

	ww := httputil.WrapWriter(w)
	w = ww

	response, err := h.ops.GetRawResponse(r.Context(), w, vehicle)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	if ww.Status() > 0 {
		return
	}

	httputil.JSONWrite(w, r, 200, response)
}

// GetTest
func (h OpenAPIHandlers) GetTest(w http.ResponseWriter, r *http.Request) {
	var err error

	logctx.AddStrToContext(r.Context(), "op", "getTest")

	var validationErrors validation.Errors

	vehicle, _, err := params.GetEnumQuery(r, "vehicle", true, NewGetTestVehicle)
	if err != nil {
		validationErrors.Add("vehicle", err)
	}

	vehicleDefault, _, err := params.GetEnumQuery(r, "vehicleDefault", true, NewGetTestVehicleDefault)
	if err != nil {
		validationErrors.Add("vehicleDefault", err)
	}

	// playerId

	playerID, _, err := params.GetUUIDQuery(r, "playerId", true)
	if err != nil {
		validationErrors.Add("playerId", err)
	}

	color, _, err := params.GetEnumQuery(r, "color", true, NewColorQuery)
	if err != nil {
		validationErrors.Add("color", err)
	}

	colorDefault, _, err := params.GetEnumQuery(r, "colorDefault", true, NewColorQueryDefault)
	if err != nil {
		validationErrors.Add("colorDefault", err)
	}

	season, _, err := params.GetEnumPath(r, "season", true, NewSeason)
	if err != nil {
		validationErrors.Add("season", err)
	}

	if validationErrors != nil {
		httputil.ErrorHandler(w, r, validationErrors.GetErr())

		return
	}

	response, err := h.ops.GetTest(r.Context(), vehicle, vehicleDefault, playerID, color, colorDefault, season)
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	httputil.JSONWrite(w, r, 200, response)
}
